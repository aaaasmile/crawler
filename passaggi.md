# Crawler
Programma che uso per ricevere una mail settimanale con i chart dei miei indici.
Nel database sqlite metto dentro tutte le info degli ISNI ed
eseguo un crawler sul sito dei chart. 
I chart vengono scaricati e inviati con una mail usando relay di invido.it.

Le credential sono nel db

Per avere cgo bisogna settare il path:

    $env:path="C:\TDM-GCC-64\bin;" + $env:path

## Nuovo sito
Il service del chart è stato aggiornato in 04.23. Non ci sono più chart
in formato gif ma svg su un sito tutto nuovo. Per salvare le figure 
in formato svg vedi la sotto directory sometest. Il crawler  github.com/gocolly/colly
non sembra in grado di scaricare html che si aggiorna in background. Per questo
ho cominciato a scaricare i chart usando github.com/chromedp (vedi somtest dir).


## Deployment
Questo programma viene lanciato tutte le settimane da un cronjob su pi3-hole
Questo è il comando che ho usato in crontab (ogni venerdì alle 18:28)
28 18 * * 5  cd /home/igors/projects/go/crawler && ./crawler.bin > /tmp/crawler.log
Per fare andare crontab -e bisogna lanciare sudo raspi-config e settare la time zone.
Dopo un reboot crontab -e funziona.

## Aggiornare il programma
Per aggiornare il programma crawler su pi3-hole basta aggiornarlo su windows e 
poi con la legacy console:
ssh pi3-hole
cd /home/igors/projects/go/crawler
git pull
go build -o crawler.bin

Per avere il db in locale dal target:
rsync -chavzP --stats igors@pi3.local:/home/igors/projects/go/crawler/chart-info.db . 
Per rimetterlo indietro:
rsync -chavzP --stats ./chart-info.db igors@pi3.local:/home/igors/projects/go/crawler/chart-info.db

Poi basta lanciare ./crawler.bin per vedere se tutto funziona a dovere.

## Email Relay su invido.it
Ho settato un service smtp di relay (https://github.com/aaaasmile/mailrelay-invido) che non è affatto male in quanto usa un account come gmx molto affidabile per l'invio delle mail usando tls (con gmail non è possibile, vedi sotto).
Per vedere come si manda la mail vedi  
D:\scratch\go-lang\mail-relay\ref\smtpd-master\client\client_example.go

Mandare le mail con il relay ha avuto delle trappole, tipo la codifica
delle apici da parte del server gmx. Questo ha distrutto in gran parte 
il formato html della mail.  
L'ho risolto codificando il contenuto della mail html in formato rfcbase64.
Da notare che la codifica di tutto il messaggio non funziona, ma si possono 
codificare solo le sezioni.

Nota che per usare il relay di invido, le credential sono nel db. Secret File json 
viene usato solo per google.


## Mandare le mail con google (idea abbandonata)

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

Però la Mail di prova, con il token manuale valido 7 giorni, funziona a meraviglia.

## Refresh Token di google (non usato, idea abbandonata)
Dopo una lettura della documentazione https://developers.google.com/identity/protocols/oauth2
risulta chiaro che un'applicazione web, ma destinata a rimanere in fase di test,
ha un refresh token valido per una sola settimana. Sei mesi se l'app è approvata, ma siccome non
è neanche web, non ha nessuna possibilità di esserlo. 
Il post del blog è bello per vedere un risultato, ma rappresenta uno scenario non reale,
in quanto non posso certo aggiornare manualmente un token nella dev-console ogni 7 giorni,
quando il report ha proprio questa scadenza.
Lo scenario di una applicazione web è quello di chiedere all'utente un'autorizzazione 
via web che ritorna al link dell'app una volta concessa.

Per un service senza web interface come questo _crawler_ non è la soluzione corretta.
Quindi proviamo ad usare un service account che manda un token JWT in cambio riceve 
un auth token da usare subito senza refresh.
La documentazione si trova su: https://developers.google.com/identity/protocols/oauth2/service-account#httprest

Purtroppo anche il Service Account non sembra avere molta fortuna senza avere un 
account aziendale. Arrivo a generare il JWt, l'access token, ma al momento di mandare 
la mail, appare questo errore abbastanza decisivo :
 _googleapi: Error 400: Precondition check failed., failedPrecondition_  
Alla fine la mia impressione è che gmail a livello gratuito non vuole garantire servizi continui
che non abbiano interazione con la pagina di gmail, anche solo per dare la conferma dell'accesso.

