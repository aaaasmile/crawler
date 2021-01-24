# Crawler
Programma che uso per ricevere una mail settimanale con i chart dei miei indici.
Nel database sqlite metto dentro tutte le info degli ISNI ed
eseguo un crawler sul sito dei chart. 
I chart vengono scaricati e inviati con una mail usando gmail e auth.

Le credential sono nel db

Per mandare la Mail con gmail ho seguito questo post:
https://medium.com/wesionary-team/sending-emails-with-go-golang-using-smtp-gmail-and-oauth2-185ee12ab306

Settare tutte le credential è davvero complesso. 
I 2 token (AuthCode e RefreshToken) li ho creati qui:
https://developers.google.com/oauthplayground
Qui poi ho avuto il problema che il token è scaduto.
Quindi l'ho rigenerato con oauthplayground. Il processo non è immediato.
Quello che bisogna fare è mettere nei settings "Use your own OAuth credentials"
cliccando questa checkbox. Lì si mette il clientid e il secret id del client.
Entrambi valori sono nel database.
Ora, con questi  settings, bisogna mettere in fondo a basso nella parte sinistra la url:
https://mail.google.com
e premere il pulsante "Authorize API" dello step 1.
Nello step 2 compare l'auth code. Premendo il pulsante "Exchange authorization code for tokens"
viene generato il Refreshtoken e il token della session. A me interessa solo
"Authtoken" e "RefreshToken" che vanno messi nel db. 
Dopo una settimana sembra che la coppia "Authtoken" e "RefreshToken" non sia più attuale.
Quindi vorrei rinnovarla nel codice, ma non ho trovato il modo se non questo metodo manuale.



Mentre le credential del Client (client id e client secret) qui:
https://console.cloud.google.com/
La parte più difficoltosa è stata la pagina di consent, dove solo alla fine ho potuto inserire un test user.
https://console.cloud.google.com/apis/credentials/consent?project=mailcharter


Però la Mail di prova funziona a meraviglia.

