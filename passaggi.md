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
Al momento ho eliminato la sezione delle immagini dal template (tag img)

## TODO
 - vedi di mettere l'immagine svg del chart nella mail. Il tag svg ha bisogno dei css.

## Deployment
Questo programma viene lanciato tutte le settimane da un cronjob su pi3-hole
Questo è il comando che ho usato in crontab (ogni venerdì alle 18:28)
28 18 * * 5  cd /home/igors/projects/go/crawler && ./crawler.bin > /tmp/crawler.log
Per fare andare crontab -e bisogna lanciare sudo raspi-config e settare la time zone.
Dopo un reboot crontab -e funziona.

## Aggiornare il programma
Per aggiornare il programma crawler su pi3-hole basta aggiornarlo su windows e 
poi con WSL:
ssh igors@pi3.local
cd /home/igors/projects/go/crawler
git pull
go build -o crawler.bin
(al momento uso il branch easyservice)

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

