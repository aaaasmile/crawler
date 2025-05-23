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
	
## Nuovo sito aprile 2023
Il service del chart è stato aggiornato in 04.23. Non ci sono più chart
in formato gif ma svg su un sito tutto nuovo. Per salvare le figure 
in formato svg vedi la sotto directory scraper. Il crawler  github.com/gocolly/colly
non sembra in grado di scaricare html che si aggiorna in background. Per questo
ho cominciato a scaricare i chart usando github.com/chromedp (vedi somtest dir).
Al momento ho eliminato la sezione delle immagini dal template (tag img)

## svg (progetto scraper)
Le nuove immagini sono in formato svg. Però hanno anche il tag class che deve essere
incluso. Nella directory "scraper", riesco a scaricare il file svg senza problemi 
(posizionarsi sul chart 6 mesi, però, non è triviale), però
quando lo apro risulta nero. 
Il processo di conversione avviene in due step. Per prima cosa uso uno scrap per eseguire
il download del file svg, che al momento viene messo in scraper/static/data/
Poi uso un http server per mostrare il file svg e fare in modo che attraverso il canvas diventi
un'immagine png. Lo style per il grafico è messo dentro al file main.css che ho trovato quando
ho salvato la pagina dal browser sul mio hard disk.
Per capire come funziona la visualizzazione del svg nel canvas sono partito dall'esempio dell'ellisse,
che dopo diverse prove ha funzionato. Per il file scaricato chart02.svg, la sua visualizzazione 
in html funziona, ma non quella nel canvas per via, credo, degli styles.

## svg to png
Come prima soluzione ho usato un canvas su un http server integrato che poi viene
scaricato in formato png (funzione saveToPngItem). A me sembra ora un overkill in quanto
la funzione takeSVGScreenshot riesce a salvare un componente della pagina in formato png
senza bisogno del download in formato svg e successiva conversione.

### svg nel canvas
Ho impiegato un po' a creare un canvas che disegni il mio svg scaricato dal sito dei chart.
Il motivo è che, nel canvas, l'immagine svg deve inglobare al suo interno gli stylesheets che
servono per mostrare l'immagine. Nel mio caso è il file main.css. Il procedimento è quello di
prendere il file svg usando document.getElementById. 
Nel mio caso è il firstchild del div id="thesvg" (nota che è una property e non una funzione).
Ora con l'elemento del DOM svg in mano, si deve inserire al primo posto il contenuto di main.css
 (basta solo questo e non tutti gli altri css) in un Dom def->style. Il mio dom svg ha ora 17 children node, anzichè 16 originali. Ora basta ricreare il sorgente xml e per questo si 
 usa (new XMLSerializer()).serializeToString(thesvg). Il sorgente xml diventa il contenuto di
 un Blob, che a sua volta viene identificato da una url (si usa DOMURL.createObjectURL() ).
 Questa url divena il sorgente dell'immagine da mostrare nel canvas (img.src = url e img.onload).  
Due aspetti non sono ancora corretti. Il primo sono i font. Nel trace del service noto

    GET requested  /svg/fonts/DINPro-Regular.woff
che signica che la url del font invece di fonts/DINPro-Regular.woff dovrebbe essere 
static/css/fonts/DINPro-Regular.woff
La seconda è di natura cosmetica, ma nella Array.prototype.forEach.call(sheets, function(sheet)
dovrebbe comparire solo main.css.
Per i font ho provato questa sequenza in main.css:
src: url('data:application/font-woff;charset=utf-8;base64,d09GRk9...');
L'ho provato per il font DINPro-Regular.woff, che è quello che carica quando viene mostrato il grafico. 
Il problema è che nel canvas non viene usato anche se è embedded. Allora ho ripristinato main.css
in quanto si carica più velocemente.

Queste le risorse usate:
- https://stackoverflow.com/questions/41340468/convert-svg-to-image-in-png
- https://stackoverflow.com/questions/41571622/how-to-include-css-style-when-converting-svg-to-png
- https://stackoverflow.com/questions/49666196/convert-svg-to-png-with-styles

### svg png nella Mail
Quando il programma riesce a scaricare il svg ed a convertirlo in png, basta che poi metta
il file nella directory data. Il nome è chart_{id}.png, dove id è la primary key del record nel db di 
stocklist (vedi la funzione buildChartListFromLastDown). Poi si tratta di ripristinare il tag img nel 
template della mail
    
    <img src="cid:{{.ImgName}}" alt="{{.CurrentPrice}}" />
Non va chiamata la funzione buildChartListFromLastDown, che comunque va ripristinata nel nome del file,
ma nella buildTheChartList, dove ho messo il TODO.
Così ho due exe che vanno in cascata. Il primo scarica i files svg e li converte in png per ogni stockprice.
Il secondo programma riceve i dati dei prezzi, aggiunge il file png scaricato del chart ed invia la mail.  

## TODO
 - vedi di mettere l'immagine svg del chart nella mail. Manca lo scraping partendo dal db. [DONE]
 - nel download dello scrap, il blocking del download deve avere un timeout. [DONE] 

## Deployment su invido
Ho fatto il deployment sul server dell'invido, dove ho aggiornato golang ed ho installato
google-chrome. Per aggiornare golang ho seguito le istruzioni della homepage di golang,
dove ho scaricato il tar, cancellata la distribuzione corrente in /usr/local/go e
scompattao il tar nel /usr/local/go (vedi anche le istruzioni del raspberry).
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



## Sezione Obsoleta pi3 hole
Non utilizzo più pi3 a malinquore, ma il fatto di dover usare ora chrome-headless
mi ha fatto abbandonare l'idea.

## Deployment (obsoleto)
Questo programma viene lanciato tutte le settimane da un cronjob su pi3-hole
Questo è il comando che ho usato in crontab (ogni venerdì alle 18:28)
28 18 * * 5  cd /home/igors/projects/go/crawler && ./crawler.bin > /tmp/crawler.log
Per fare andare crontab -e bisogna lanciare sudo raspi-config e settare la time zone.
Dopo un reboot crontab -e funziona. Ad un certo però, su pi3-hole crontab non ha più funzionato.
Vedi il file di readme-pihole di per come ho risolto, ma ho dovuto usare un'alternativa a crontab.
Per questo ho usato Anacron, che però, mi manda l'email giovedi sera anziché il venerdì sera.

## Aggiornare il programma (nota che non uso più pi3 per via dello scraper)
Per aggiornare il programma crawler su pi3-hole basta aggiornarlo su windows e 
poi con WSL:

    ssh igors@pi3.local
    cd /home/igors/projects/go/crawler
    git pull
    git mod tidy
    go build -o crawler.bin

Per avere il db in locale dal target:
rsync -chavzP --stats igors@pi3.local:/home/igors/projects/go/crawler/chart-info.db . 
Per rimetterlo indietro:
rsync -chavzP --stats ./chart-info.db igors@pi3.local:/home/igors/projects/go/crawler/chart-info.db

Poi basta lanciare ./crawler.bin per vedere se tutto funziona a dovere.
Ho dovuto fare un aggiornamento di go alla versione 1.21 e qui ho avuto qualche problema.
La versione da installare è la arm6l e non la arm64. In più ero su un branch che poi ho cancellato in remoto.
Quando poi ricompilo il programma con go nuovo di pacca, impiega molto tempo perché ricompila con cc sqlite3.
In più pi3 è molto lento, ma funziona.
Un altro problema l'ho avuto con x/clipboard, che poi ho tolto per evitare problemi con 
    
    #include <X11/Xlib.h>.
Per verdere la lista dei branch:

    git branch
    git checkout main
    git pull


