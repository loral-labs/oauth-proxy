# Loral

Loral is an integrated application authorization service that provides a unified API for various services like Google Drive, Gmail, Notion, and more. 

## How it Works

Instead of having to configure your application with each individual service and manage multiple access tokens, you can simply use Loral. 

For example, if you need to call `https://api.notion.com/api_action`, you would instead call `https://api.loral.dev/notion/execute/api_action` with your Loral access token. 

As long as the user has authorized Loral to access their Notion account, and your application is authorized to access their Loral account, everything is all set!

## Benefits

- Simplified Authorization: No need to manage multiple access tokens for different services.
- Unified API: One API to interact with all the services.
- Easy Integration: Users just need to authorize your application to their Loral and you automatically get access to all of the services they have authorized Loral to.

Start building with Loral today and simplify your application's integration with various services.