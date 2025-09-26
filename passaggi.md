# Crawler
Programma che uso per ricevere una mail settimanale con i chart dei miei indici.
Nel database sqlite metto dentro tutte le info degli ISNI ed
eseguo un crawler sul sito dei chart. 
I chart vengono scaricati e inviati con una mail usando relay di invido.it.

Le credential sono nel db

Per avere cgo (richiesto da go-sqlite3) bisogna settare il path:

    $env:path="C:\TDM-GCC-64\bin;" + $env:path

Sul mio mini-k7 per effettuare un build devo usare:

	go build -buildvcs=false

Su server dell'invido devo lanciare due programmi:

    - scraper.bin
    - crawler.bin

Scarper mi server per scaricare le immagini(usa l'installazione locale di google-chrome), 
mentre crawler.bin manda la mail inglobando le immagini precedentemente scaricate.
crawler.bin è responsabile anche del parsing dei valori delle quortazioni.
	
## Nuovo sito aprile 2023
Il service del chart è stato aggiornato in 04.23. Non ci sono più chart
in formato gif ma svg su un sito tutto nuovo. Per salvare le figure 
in formato svg vedi la sotto directory scraper. Il crawler  github.com/gocolly/colly
non sembra in grado di scaricare html che si aggiorna in background. Per questo
ho cominciato a scaricare i chart usando github.com/chromedp (vedi somtest dir).
Al momento ho eliminato la sezione delle immagini dal template (tag img)

## svg (progetto scraper)
Le nuove immagini sono in formato svg. Per salvarle in png uso la funzionalità takeSVGScreenshot
senza web server con canvas.

## svg to png
La funzione takeSVGScreenshot riesce a salvare un componente della pagina in formato png
senza bisogno del download in formato svg e successiva conversione.

### svg png nella Mail
Quando il programma riesce a scaricare il svg in png, basta che poi metta
il file nella directory data. Il nome è chart_{id}.png, dove id è la primary key del record nel db di 
stocklist (vedi la funzione buildChartListFromLastDown). Poi si tratta di ripristinare il tag img nel 
template della mail
    
    <img src="cid:{{.ImgName}}" alt="{{.CurrentPrice}}" />
Non va chiamata la funzione buildChartListFromLastDown, che comunque va ripristinata nel nome del file,
ma nella buildTheChartList, dove ho messo il TODO.
Così ho due exe che vanno in cascata. Il primo scarica i files svg e li converte in png per ogni stockprice.
Il secondo programma riceve i dati dei prezzi, aggiunge il file png scaricato del chart ed invia la mail.  

## Deployment su invido
Ho fatto il deployment sul server dell'invido, dove ho aggiornato golang ed ho installato
google-chrome. Per aggiornare golang ho seguito le istruzioni della homepage di golang,
dove ho scaricato il tar, cancellata la distribuzione corrente in /usr/local/go e
scompatto il tar nel /usr/local/go (vedi anche le istruzioni del raspberry).
Qui ho avuto un problema col click del cambio scala del grafico (6 mesi). La ragione è dovuta
al popup dei cookies, che mi compare solo su alcune distribuzioni di WSL e invido, ma non su Windows.
Lo screenshot del contenuto del chart mi ha mostrato il problema. Il selector dei pulsanti
dei cookies non ha funzionato, per cui ho usato un click con coordinate assolute sulla view 1920x1080.
Per lo sviluppo su invido ho usato Code in collegamento ssh 
(per la connessione si clicca in basso a sinistra) che è molto utile. 

### Autostart
Ho usato crontab -e con queste due lineee

    28 18 * * 5 cd /home/igor/app/go/crawler/scraper && ./scraper.bin > /tmp/scraper.log
    45 18 * * 5 cd /home/igor/app/go/crawler && ./crawler.bin > /tmp/crawler.log

## Scraper su Ubuntu di invido
Ho dovuto installare google-chrome in quanto ho ricevuto il seguente errore:
 
    "google-chrome": executable file not found in $PATH
Ho installato google-chrome con la seguente sequenza:

    cd tmp
    wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
    sudo dpkg -i google-chrome-stable_current_amd64.deb; apt-get -fy install
    sudo apt-get --fix-broken install
Se ci sono dei problemi con le dipendenze (esempio in ubunut 24.04):

    sudo apt --fix-broken install

Qui https://github.com/geziyor/geziyor/issues/27 viene spiegato il problema.
Un altro link utile è: https://github.com/Zenika/alpine-chrome
Il prompt che ottengo:

    google-chrome --version
    Google Chrome 120.0.6099.109

## Email Relay su invido
Ho settato un service smtp di relay (https://github.com/aaaasmile/mailrelay-invido) che non è affatto male in quanto usa un account come gmx molto affidabile per l'invio delle mail usando tls (con gmail non ci sono riuscito, vedi passaggi_gmail.md).
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

## Problemi
Mi è comparso un errore del genere:

    ERROR: could not unmarshal event: parse error: expected string near offset 1081 of 'cookiePart...'
La soluzione è stata quella di effettuare un upgrade di chromedp

     go get -u github.com/chromedp/chromedp
Altro errore:

    [scrapItem] error on chromedp.Run context deadline exceeded
Questo si ha quando la query su un nodo non va a buon fine. Il contesto si esaurisce
e non può più essere usato. Per nodi che sono opzionali, occorre due contesti.

Nel Log questo:

    2025/09/26 22:20:00 ERROR: could not unmarshal event: json: cannot unmarshal JSON string into Go network.IPAddressSpace within "/clientSecurityState/initiatorIPAddressSpace": unknown IPAddressSpace value: Loopback
sembra dovuto a qualche oscuro log di chromedp che esegue il download su localhost.
Ho ignorato l'errore in quanto il download funziona.
