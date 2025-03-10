use anyhow::Result;
use keyring::Entry;
use oauth2::reqwest::Client;
use oauth2::{basic::BasicClient, basic::BasicTokenType, TokenResponse};
use oauth2::{
    reqwest, AccessToken, EmptyExtraTokenFields, StandardRevocableToken, StandardTokenResponse,
};
use oauth2::{
    AuthUrl, AuthorizationCode, ClientId, ClientSecret, CsrfToken, PkceCodeChallenge, RedirectUrl,
    RevocationUrl, Scope, TokenUrl,
};
use std::env;
use std::io::{BufRead, BufReader, Write};
use std::net::TcpListener;
use tracing::{info, warn};
use url::Url;

/// Google OAuth for google tasks
pub struct GoogleOAuth {
    connected: bool,
    pub access_token: String,
    pub refresh_token: String,
}

impl GoogleOAuth {
    pub fn new() -> Self {
        Self {
            connected: false,
            access_token: "".into(),
            refresh_token: "".into(),
        }
    }

    pub async fn set_tokens(
        &self,
        token_response: StandardTokenResponse<EmptyExtraTokenFields, BasicTokenType>,
    ) -> Result<()> {
        Entry::new("tasked", "access-token")?
            .set_secret(token_response.access_token().secret().as_bytes())?;

        Entry::new("tasked", "refresh-token")?
            .set_secret(token_response.refresh_token().unwrap().secret().as_bytes())?;

        Ok(())
    }

    pub async fn get_tokens(&mut self) -> Result<String> {
        self.access_token = String::from_utf8(Entry::new("tasked", "access-token")?.get_secret()?)?;
        self.refresh_token =
            String::from_utf8(Entry::new("tasked", "refresh-token")?.get_secret()?)?;
        Ok(self.access_token.clone())
    }

    pub fn get_http_client(&self) -> Result<Client> {
        let http_client = reqwest::ClientBuilder::new()
            // Following redirects opens the client up to SSRF vulnerabilities.
            .redirect(reqwest::redirect::Policy::none())
            .build()
            .expect("Client should build");

        Ok(http_client)
    }

    pub fn create_google_oauth_client(self) -> Result<BasicClient> {
        unimplemented!()
    }

    pub async fn is_connected(&mut self) -> Result<bool> {
        // check if the user is already signed in
        self.get_tokens().await?;

        Ok(self.connected)
    }

    pub async fn sign_in(mut self) -> Result<()> {
        info!("Login into google tasks");
        if self.is_connected().await.unwrap_or(false) {
            warn!("Already connected to google tasks, Skipping sign in");
            return Ok(());
        }

        let google_client_id = ClientId::new(
            env::var("GOOGLE_CLIENT_ID")
                .expect("Missing the GOOGLE_CLIENT_ID environment variable."),
        );
        let google_client_secret = ClientSecret::new(
            env::var("GOOGLE_CLIENT_SECRET")
                .expect("Missing the GOOGLE_CLIENT_SECRET environment variable."),
        );
        let auth_url = AuthUrl::new("https://accounts.google.com/o/oauth2/v2/auth".to_string())
            .expect("Invalid authorization endpoint URL");
        let token_url = TokenUrl::new("https://www.googleapis.com/oauth2/v3/token".to_string())
            .expect("Invalid token endpoint URL");

        // Set up the config for the Google OAuth2 process.
        let client = BasicClient::new(google_client_id)
            .set_client_secret(google_client_secret)
            .set_auth_uri(auth_url)
            .set_token_uri(token_url)
            // This example will be running its own server at localhost:8080.
            // See below for the server implementation.
            .set_redirect_uri(
                RedirectUrl::new("http://localhost:8080".to_string())
                    .expect("Invalid redirect URL"),
            )
            // Google supports OAuth 2.0 Token Revocation (RFC-7009)
            .set_revocation_url(
                RevocationUrl::new("https://oauth2.googleapis.com/revoke".to_string())
                    .expect("Invalid revocation endpoint URL"),
            );

        let http_client = self.get_http_client()?;

        // Google supports Proof Key for Code Exchange (PKCE - https://oauth.net/2/pkce/).
        // Create a PKCE code verifier and SHA-256 encode it as a code challenge.
        let (pkce_code_challenge, pkce_code_verifier) = PkceCodeChallenge::new_random_sha256();

        // Generate the authorization URL to which we'll redirect the user.
        let (authorize_url, csrf_state) = client
            .authorize_url(CsrfToken::new_random)
            // This example is requesting access to the "calendar" features and the user's profile.
            .add_scope(Scope::new(
                "https://www.googleapis.com/auth/tasks".to_string(),
            ))
            .add_scope(Scope::new(
                "https://www.googleapis.com/auth/tasks.readonly".to_string(),
            ))
            .set_pkce_challenge(pkce_code_challenge)
            .url();

        println!("Open this URL in your browser:\n{authorize_url}\n");

        open::that(authorize_url.to_string()).unwrap();

        let (code, state) = {
            // A very naive implementation of the redirect server.
            let listener = TcpListener::bind("127.0.0.1:8080").unwrap();

            // The server will terminate itself after collecting the first code.
            let Some(mut stream) = listener.incoming().flatten().next() else {
                panic!("listener terminated without accepting a connection");
            };

            let mut reader = BufReader::new(&stream);

            let mut request_line = String::new();
            reader.read_line(&mut request_line).unwrap();

            let redirect_url = request_line.split_whitespace().nth(1).unwrap();
            let url = Url::parse(&("http://localhost".to_string() + redirect_url)).unwrap();

            let code = url
                .query_pairs()
                .find(|(key, _)| key == "code")
                .map(|(_, code)| AuthorizationCode::new(code.into_owned()))
                .unwrap();

            let state = url
                .query_pairs()
                .find(|(key, _)| key == "state")
                .map(|(_, state)| CsrfToken::new(state.into_owned()))
                .unwrap();

            let message = "Go back to your terminal :)";
            let response = format!(
                "HTTP/1.1 200 OK\r\ncontent-length: {}\r\n\r\n{}",
                message.len(),
                message
            );
            stream.write_all(response.as_bytes()).unwrap();

            (code, state)
        };

        println!("Google returned the following code:\n{}\n", code.secret());
        println!(
            "Google returned the following state:\n{} (expected `{}`)\n",
            state.secret(),
            csrf_state.secret()
        );

        // Exchange the code with a token.
        let token_response = client
            .exchange_code(code)
            .set_pkce_verifier(pkce_code_verifier)
            .request_async(&http_client)
            .await?;
        self.set_tokens(token_response).await?;

        self.connected = true;
        Ok(())
    }

    pub async fn sign_out(&mut self) -> Result<()> {
        let google_client_id = ClientId::new(
            env::var("GOOGLE_CLIENT_ID")
                .expect("Missing the GOOGLE_CLIENT_ID environment variable."),
        );
        let google_client_secret = ClientSecret::new(
            env::var("GOOGLE_CLIENT_SECRET")
                .expect("Missing the GOOGLE_CLIENT_SECRET environment variable."),
        );
        let auth_url = AuthUrl::new("https://accounts.google.com/o/oauth2/v2/auth".to_string())
            .expect("Invalid authorization endpoint URL");
        let token_url = TokenUrl::new("https://www.googleapis.com/oauth2/v3/token".to_string())
            .expect("Invalid token endpoint URL");

        // Set up the config for the Google OAuth2 process.
        let client = BasicClient::new(google_client_id)
            .set_client_secret(google_client_secret)
            .set_auth_uri(auth_url)
            .set_token_uri(token_url)
            // This example will be running its own server at localhost:8080.
            // See below for the server implementation.
            .set_redirect_uri(
                RedirectUrl::new("http://localhost:8080".to_string())
                    .expect("Invalid redirect URL"),
            )
            // Google supports OAuth 2.0 Token Revocation (RFC-7009)
            .set_revocation_url(
                RevocationUrl::new("https://oauth2.googleapis.com/revoke".to_string())
                    .expect("Invalid revocation endpoint URL"),
            );

        self.get_tokens().await?;
        let token_to_revoke: StandardRevocableToken =
            AccessToken::new(self.access_token.clone()).into();

        client
            .revoke_token(token_to_revoke)?
            .request_async(&self.get_http_client()?)
            .await
            .expect("Failed to revoke token");
        Ok(())
    }
}
