# Proxy to localhost

You need to use tunneling to access your local server from the internet. You can use [ngrok](https://ngrok.com/) for this purpose.

Then set the env PROXY_URL to the ngrok url.

Every request to the proxy url will be forwarded to your local server.
