# Loral

Loral is an integrated application authorization service that provides a unified API for various services like Google Drive, Gmail, Notion, and more.

## How it Works

Instead of having to configure your application with each individual service and manage multiple access tokens, you can simply use Loral.

For example, if you need to call `https://api.notion.com/api_action`, you would instead call `https://api.loral.dev/notion/v1/api_action` with your Loral access token.

As long as the user has authorized Loral to access their Notion account, and your application is authorized to access their Loral account, everything is all set!

## Benefits

- Simplified Authorization: No need to manage multiple access tokens for different services.
- Unified API: One API to interact with all the services.
- Easy Integration: Users just need to authorize your application to their Loral and you automatically get access to all of the services they have authorized Loral to in the past.

Start building with Loral today and simplify your application's integration with various services.

## Usage

### Authorization

Our authorization is the standard OAuth 2.0 flow:

1. First register your application by going to `loral.dev` and you will receive a `LORAL_CLIENT_ID` AND `LORAL_CLIENT_SECRET`

2. Next run an authorization request as shown below:

```
curl -X GET \
https://auth.loral.dev/oauth2/auth?scope={{SCOPE}}&response_type=code&client_id={{LORAL_CLIENT_ID}}&redirect_uri={{REDIRECT_URI}}&state={{STATE}} \
  -H 'Cache-Control: no-cache' \
  -H 'Content-Type: application/x-www-form-urlencoded'
```

You will receive a response containing the `AUTHORIZATION_CODE` and `STATE` as query parameters to your `REDIRECT_URI`.

3. Next run a token request as shown below:

```
curl -X POST \
  'https://auth.loral.dev/oauth2/token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -H 'Authorization: Basic {{base64(LORAL_CLIENT_ID:LORAL_CLIENT_SECRET)}}' \
  -d 'grant_type=authorization_code&code={{AUTHORIZATION_CODE}}&redirect_uri={{REDIRECT_URI}}'
```

You will receive a response JSON containing the keys: `access_token`, `refresh_token`, `expires_in` and `scope`.

Now you have a **Loral access token** that you can use for all of your requests for any application within the `scope` variable.

4. To refresh your token, you can use a refresh request as shown below:

```
curl -X POST \
  'https://auth.loral.dev/oauth2/token' \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -H 'Authorization: Basic {{base64(LORAL_CLIENT_ID:LORAL_CLIENT_SECRET)}}' \
  -d 'grant_type=refresh_token&refresh_token={{REFRESH_TOKEN}}'
```

You will then receive a response JSON containing the keys same keys `access_token`, `refresh_token`, `expires_in` and `scope`.

### Execution

For executing APIs, first please refer to the `providers.json` file in our repository. This will show whether or not the server URL you are trying to access is has been indexed by Loral. If you do find your server URL as a key then find the provider name corresponding to that url.

Then instead of sending your request to `{serverURL}/{path}` you should instead send your request to `https://api.loral.dev/{providerName}/v1/{path}` with the same parameters, headers and request body. The only difference should be that you must set the header `"Authorization": "Bearer {LORAL_ACCESS_TOKEN}"` and we will return the same response.
