# Loral

Loral is an integrated application authorization service that provides a unified API for various services like Google Drive, Gmail, Notion, and more.

## How it Works

Instead of having to configure your application with each individual service and manage multiple access tokens, you can simply use Loral.

For example, if you need to call `https://api.notion.com/api_action`, you would instead call `https://api.loral.dev/notion/execute/api_action` with your Loral access token.

As long as the user has authorized Loral to access their Notion account, and your application is authorized to access their Loral account, everything is all set!

## Benefits

- Simplified Authorization: No need to manage multiple access tokens for different services.
- Unified API: One API to interact with all the services.
- Easy Integration: Users just need to authorize your application to their Loral and you automatically get access to all of the services they have authorized Loral to in the past.

Start building with Loral today and simplify your application's integration with various services.

## Usage

### Authorization

Our authorization is the standard OAuth 2.0 flow:

### 1. First register and configure your application using the following endpoints and `api.loral.dev` as the hostname. Make sure to save the ClientId and ClientSecret returned to you after creating a Client — these will only be shown once. Endpoints 2-4 are if you need to make future edit — making edits requires the ClientId and ClientSecret.

1.  Create OAuth Client

- **URI:** `/client/create`
- **Method:** POST
- **Input:**
  - **Body** (`application/json`):
    ```json
    {
      "name": "string",
      "redirect_uris": ["string"],
      "scopes": ["string"] // a list of providers your app needs access to, ie. ["google", "kroger"]
    }
    ```
- **Output:**
  - **Success (200 OK)** (`application/json`):
    ```json
    {
      "id": "string",
      "secret": "string"
    }
    ```
  - **Error (400 Bad Request / 500 Internal Server Error)**: Error message as plain text.

2. Edit OAuth Client Name

- **URI:** `/client/edit/name`
- **Method:** POST
- **Input:**
  - **Body** (`application/json`):
    ```json
    {
      "id": "string",
      "secret": "string",
      "name": "string"
    }
    ```
- **Output:**
  - **Success (200 OK)**: No body, indicates successful operation.
  - **Error (400 Bad Request / 500 Internal Server Error)**: Error message as plain text.

3. Edit OAuth Client Scope

- **URI:** `/client/edit/scope`
- **Method:** POST
- **Input:**
  - **Body** (`application/json`):
    ```json
    {
      "id": "string",
      "secret": "string",
      "name": "string",
      "add": boolean
    }
    ```
- **Output:**
  - **Success (200 OK)**: No body, indicates successful operation.
  - **Error (400 Bad Request / 500 Internal Server Error)**: Error message as plain text.

4. Edit OAuth Client Redirect URIs

- **URI:** `/client/edit/redirectUris`
- **Method:** POST
- **Input:**
  - **Body** (`application/json`):
    ```json
    {
      "id": "string",
      "secret": "string",
      "uris": ["string"]
    }
    ```
- **Output:**
  - **Success (200 OK)**: No body, indicates successful operation.
  - **Error (400 Bad Request / 500 Internal Server Error)**: Error message as plain text.

### 2. Next request an authorization code:

Note that these endpoints use `auth.loral.dev` as the hostname.

You'll need to redirect your user from the frontend of your application to the following URL. Fill in the URL parameters as appropriate.

The `redirect_uri` should be a location on your app where the user will be redirected after successfully completing the flow/authorizing your client to access Loral. For example, it may be `http://example.com/api/callback`. It must be one of the redirect uris you registered as part of step one.

We recommend setting the state parameter to a stringified JSON like the one below

```
{
  csrf: crypto.randomBytes(16).toString("hex"), // random nonce for CSRF protection, verify this in the callback
  userId: String(user.id), // to be used in the callback
}
```

The `userId` should be a unique user identifier from YOUR app/DB. This way, you can associate the authorization code returned in the callback with one of your users.

```
curl -X GET \
https://auth.loral.dev/oauth2/auth?scope={{SCOPE}}&response_type=code&client_id={{LORAL_CLIENT_ID}}&redirect_uri={{REDIRECT_URI}}&state={{STATE}} \
 -H 'Cache-Control: no-cache' \
 -H 'Content-Type: application/x-www-form-urlencoded'
```

You will receive a callback containing the `AUTHORIZATION_CODE` and `STATE` as query parameters at your `REDIRECT_URI`.

### 3. Next exchange the code for access + refresh tokens:

```
curl -X POST \
 'https://auth.loral.dev/oauth2/token' \
 -H 'Content-Type: application/x-www-form-urlencoded' \
 -H 'Authorization: Basic {{base64(LORAL_CLIENT_ID:LORAL_CLIENT_SECRET)}}' \
 -d 'grant_type=authorization_code&code={{AUTHORIZATION_CODE}}&redirect_uri={{REDIRECT_URI}}'
```

Await the response. It will contain a JSON with the keys: `access_token`, `refresh_token`, `expires_in` and `scope`. Save and associate these with the user they were intended for.

Now you have a **Loral access token** that you can use for all of your requests for any application within the `scope` variable. The access token will expire every `expires_in` milliseconds, after which you'll have to exchange your refresh token for a fresh set of keys (see the step below).

### 4. To exchange your refresh token:

```
curl -X POST \
 'https://auth.loral.dev/oauth2/token' \
 -H 'Content-Type: application/x-www-form-urlencoded' \
 -H 'Authorization: Basic {{base64(LORAL_CLIENT_ID:LORAL_CLIENT_SECRET)}}' \
 -d 'grant_type=refresh_token&refresh_token={{REFRESH_TOKEN}}'
```

You will then receive a response JSON containing a new set of keys `access_token`, `refresh_token`, `expires_in` and `scope`, make sure you save these as the updated values in your DB.

### 5. Verify what apps you can access

The user may not have authorized Loral to access all the requested apps in scope. You can check which apps you can access by introspecting the token with the following endpoint.

```
curl -X GET \
 'https://api.loral.dev/auth/introspect' \
 -H 'Authorization: Bearer {LORAL_ACCESS_TOKEN}' \
```

This returns a JSON of provider:boolean entries, i.e.

```
"google": true,
"kroger": false,
```

For any unauthenticated providers/`false` values you would like to prompt the user to grant access to, call the following endpoint

```
curl -X GET \
 'https://api.loral.dev/{PROVIDER}/auth' \
 -H 'Authorization: Bearer {LORAL_ACCESS_TOKEN}' \
 -d 'redirect_uri={{REDIRECT_URI}} \
```

This returns a URL which you can redirect your user to in order to gain access to the provider. After they have granted access, the user will be redirected back to the `REDIRECT_URI` (probably a page on your site).

### Execution

For executing APIs, first please refer to the `./internal/config/config.go` file in our repository. This will show whether or not the server URL you are trying to access is has been indexed by Loral. If you do find your server URL as a key then find the provider name corresponding to that url.

Then instead of sending your request to `{serverURL}/{path}` you should instead send your request to `https://api.loral.dev/{providerName}/execute/{path}` with the same parameters, headers and request body. The only difference should be that you must set the header `"Authorization": "Bearer {LORAL_ACCESS_TOKEN}"` and we will return the same response.
