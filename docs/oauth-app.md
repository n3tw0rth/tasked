# tasked's public OAuth client

`tasked login` signs in with a built-in OAuth client so users never touch
Google Cloud. This is the maintainer checklist for that client. It lives in
a Google Cloud project used only for tasked, never a personal one: every user
signs in through it and shares its quota.

## One-time setup

1. Create a new Google Cloud project (e.g. "tasked") and enable the
   **Google Tasks API**.
2. Google Auth Platform → **Branding**: app name, support email, app homepage,
   and privacy policy. The privacy policy must be on the same domain as the
   homepage, and the domain must be verified in Google Search Console with an
   account that is Owner or Editor on the project.
3. **Data Access**: add only the scopes the code requests
   (`internal/auth/oauth.go`): `.../auth/tasks` and `.../auth/userinfo.email`.
4. **Audience**: user type **External**, then **Publish app**. Publishing
   also stops the 7-day refresh-token expiry that applies in Testing.
5. **Clients**: create an OAuth client ID of type **Desktop app**.
6. Paste its ID and secret into `builtinClient` in
   `internal/auth/builtin.go`. Committing them is intended: Google treats
   desktop-app client secrets as public, and PKCE protects the sign-in.

## Verification

The Tasks scope needs Google's verification before the app can serve more
than 100 new users; until then users see a "Google hasn't verified this app"
warning they can click through. Submit from **Verification Center** with:

- a justification for the Tasks scope (tasked reads and edits the user's
  tasks; `tasks.readonly` is too narrow because it creates, edits, and
  deletes tasks)
- an unlisted YouTube demo video, in English, showing `tasked login`, the
  consent screen with the app name, the OAuth client ID in the browser
  address bar, and commands that use the scope (`add`, `ls`, `done`)

Google's stated review time for sensitive scopes is 3–5 business days.

## Operating it

- Quota: 50,000 Tasks API requests per day for the whole project by default;
  watch usage in the console and request an increase before it's needed.
- Changing the client ID invalidates existing sign-ins: users get a
  "made with a different OAuth client" error and must run `tasked login`.
  Rotating only the secret (Clients → add secret, then disable the old one)
  keeps them.
