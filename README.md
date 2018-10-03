# Slack-esque
 Built a web server in Go and a microservice in Node.js with the following features and capabilites:
- Page Summary: Shows preview of a URL by extracting its meta tags
- An Infrastructure from Code using Terraform to host the dockerized API server and client server in Digital Ocean
- Track sessions using a Redis database
- Authenticate and store user information in MySQL and PostgreSQL
- Allow users to upload custom profile pictures
- Block repeated failed sign-ins
- Forgot Password: Sends an email with an base64-encoded crypto-random code that allows to reset password
- Search: Allows to search for other users based on user name, first name and last name (using trie)
- Node.js microservice for channels (public and private) and messages
- Real time notifications to relevant users for channel and message events using Websockets and RabbitMQ
- Add Emoji reactions to messages
- Star/Favorite messages
- Attach media files to messages