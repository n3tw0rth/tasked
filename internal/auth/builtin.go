package auth

// builtinClient is tasked's own public OAuth client, so users can just run
// `tasked login` without a Google Cloud account. It lives in a Google Cloud
// project dedicated to tasked (published and verified; see
// docs/oauth-app.md), not in anyone's personal project.
//
// The secret is committed on purpose: Google treats desktop-app client
// secrets as public, and PKCE protects the sign-in. Forks should register
// their own client rather than reuse this one.
//
// While these are empty, users must set TASKED_CLIENT_ID and
// TASKED_CLIENT_SECRET to their own client.
var builtinClient = Client{
	ID:     "",
	Secret: "",
	Source: "built-in",
}
