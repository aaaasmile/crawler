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
Mentre le credential del Client (client id e client secret) qui:
https://console.cloud.google.com/
La parte più difficoltosa è stata la pagina di consent, dove solo alla fine ho potuto inserire un test user.
https://console.cloud.google.com/apis/credentials/consent?project=mailcharter


Però la Mail di prova funziona a meraviglia.
