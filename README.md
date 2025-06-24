## Serveur de l'application MyWeed

> ### Pour lancer le serveur
> ### Assurez vous d'avoir le .env suivant à la racine du projet après l'avoir cloné :

```
CLIENT_URL=http://localhost:3000

MAIL_HOST=sandbox.smtp.mailtrap.io
MAIL_PORT=2525
MAIL_USERNAME=1dabae1e04b027
MAIL_PASSWORD=a5ab5ae5e77fc0    
MAIL_FROM=no-reply@myweed.com
MAIL_CONTENT_TYPE=text/html
```

L'host, le port, le username et le password correspondent aux identifiant mailtrap (serveur de test de mail pour le développement)

Afin de lancer le server :
```
go get
go run main.go
```