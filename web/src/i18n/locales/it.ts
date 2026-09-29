const it = {
  translation: {
    feedback: {
      enabled: '{{name}} attivato',
      disabled: '{{name}} disattivato',
      failed: 'La richiesta non è riuscita. Riprova.',
      network: 'Impossibile raggiungere il dispositivo. Controlla la connessione e riprova.',
      saved: 'Salvato',
      timeout: 'Il dispositivo ha impiegato troppo a rispondere. Riprova.'
    },
    common: {
      copy: 'Copia',
      copied: 'Copiato',
      copyFailed: 'Copia non riuscita. Seleziona il testo e copialo a mano.',
      notUpdating: "Non aggiornato: l'ultimo aggiornamento non è riuscito.",
      off: 'Spento',
      running: 'In esecuzione',
      save: 'Salva',
      cancel: 'Annulla',
      delete: 'Elimina',
      remove: 'Rimuovi'
    },
    head: {
      desktop: 'Desktop Remoto',
      login: 'Accesso',
      changePassword: 'Cambia Password',
      terminal: 'Terminale',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Password cambiata. Accedi con la nuova password.',
      cookieRejected:
        'Il browser si è rifiutato di salvare la sessione. Un cookie lasciato da una precedente sessione HTTPS non può essere sostituito tramite http non cifrato. Cancella i cookie di questo indirizzo, oppure apri una finestra privata, e accedi di nuovo.',
      login: 'Accesso',
      placeholderUsername: 'Inserisci il nome utente',
      placeholderPassword: 'Inserisci la password',
      placeholderCurrentPassword: 'Password attuale',
      placeholderPassword2: 'Inserisci nuovamente la password',
      noEmptyUsername: 'Il nome utente non può essere vuoto',
      noEmptyPassword: 'La password non può essere vuota',
      passwordLength: 'La password deve contenere da 8 a 72 caratteri',
      noAccount:
        'Impossibile ottenere informazioni utente, aggiorna la pagina o reimposta la password',
      invalidUser: 'Nome utente o password non validi',
      locked: 'Troppi accessi, riprova più tardi',
      globalLocked: 'Sistema sotto protezione, riprova più tardi',
      error: 'Errore imprevisto',
      invalidCurrentPassword: 'La password attuale non è corretta',
      changePassword: 'Cambia Password',
      changePasswordDesc:
        'Per la sicurezza del tuo dispositivo, modifica la password di accesso web.',
      differentPassword: 'Le password non corrispondono',
      illegalUsername: 'Il nome utente contiene caratteri non validi',
      illegalPassword: 'La password contiene caratteri non validi',
      forgetPassword: 'Hai dimenticato la password',
      ok: 'Ok',
      cancel: 'Annulla',
      loginButtonText: 'Accedi',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the IronKVM for 10 seconds.',
        reset3: 'Account web predefinito:',
        reset4: 'Account SSH predefinito:',
        change1: 'Tieni presente che questa azione modificherà le seguenti password:',
        change2: 'Password di accesso web',
        change3: 'Password root di sistema (password di accesso SSH)',
        change4: 'Per reimpostare le password, tieni premuto il pulsante BOOT sul IronKVM.',
        resetDocs: "Per i passaggi dettagliati, consulta la documentazione dell'hardware:",
        hardwareDocs: 'Wiki di Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Configura il Wi-Fi per IronKVM',
      success: 'Please check the network status of IronKVM and visit the new IP address.',
      failed: 'Operazione non riuscita, riprova.',
      invalidMode:
        'La modalità corrente non supporta la configurazione di rete. Vai al tuo dispositivo e abilita la modalità di configurazione Wi-Fi.',
      confirmBtn: 'Ok',
      finishBtn: 'Completato',
      ap: {
        authTitle: 'Autenticazione richiesta',
        authDescription: 'Inserisci la password AP per continuare',
        authFailed: 'Password AP non valida',
        passPlaceholder: 'AP password',
        verifyBtn: 'Verifica'
      },
      ssidRequired: 'Inserisci il nome della rete, fino a 32 caratteri',
      passwordLength: 'La password è di 8-63 caratteri. Lasciala vuota per una rete aperta.',
      passwordOptional: 'Password (vuota per una rete aperta)',
      lost: "La scheda ha smesso di rispondere. Potrebbe essersi collegata alla rete e aver chiuso l'hotspot di configurazione. Se l'hotspot ricompare, la connessione non è riuscita: ricollegati e riprova.",
      done: 'Configurazione completata. Ricollega questo dispositivo alla tua rete abituale e apri la scheda al suo nuovo indirizzo.'
    },
    screen: {
      viewOnly: 'Solo visione',
      viewOnlyTip:
        "Questa scheda smette di inviare tastiera e mouse all'host. Script, jiggler del mouse e altri spettatori non sono interessati.",
      viewOnlyOff: 'Disattiva solo visione',
      viewOnlyBlocked: "Solo visione è attivo, non è stato inviato nulla all'host",
      pauseHidden: 'Pausa con scheda nascosta',
      pauseHiddenTip:
        'Ferma video e audio pochi secondi dopo che questa scheda viene nascosta e li riavvia al tuo ritorno.',
      screenshot: 'Screenshot',
      screenshotTip: "Salva lo schermo dell'host come PNG alla piena dimensione di acquisizione.",
      screenshotFailed: 'Screenshot non riuscito',
      stream: {
        ok: 'immagine OK',
        noSignal: 'nessun segnale',
        failed: 'stream non riuscito'
      },
      codecNoWebrtcHevc: 'Questo browser non può ricevere H.265 tramite WebRTC',
      codecNoHevc: 'Questo browser non può decodificare H.265',
      codecNote:
        'La scheda ha un solo encoder, quindi questo cambia lo stream per tutti gli spettatori. Riconnettiti per applicarlo a una sessione WebRTC in corso.',
      codec: 'Codec',
      updateFailed: "L'impostazione non è stata applicata",
      scale: 'Scala',
      title: 'Schermo',
      video: 'Modalità video',
      videoDirectTips:
        'Abilita HTTPS in "Impostazioni > Dispositivo" per utilizzare questa modalità',
      resolution: 'Risoluzione',
      ocr: {
        title: 'Leggi testo (OCR)',
        tips: 'Il testo viene riconosciuto in questo browser. Puoi correggerlo prima di copiarlo.',
        hint: 'Trascina sul testo da leggere. Premi Esc per annullare.',
        noPicture: 'Attendi il video, poi trascina sul testo da leggere.',
        cancel: 'Annulla',
        language: 'Lingua',
        languages: {
          eng: 'Inglese'
        },
        preview: 'Area selezionata',
        capturing: 'Acquisizione dello schermo...',
        loading: 'Caricamento del riconoscimento del testo...',
        recognizing: 'Lettura del testo...',
        noText: "Nessun testo trovato nell'area selezionata.",
        copy: 'Copia',
        copied: 'Copiato negli appunti',
        copyFailed: 'Impossibile copiare negli appunti',
        selectAgain: 'Seleziona di nuovo',
        unsupported:
          'Questo browser non può eseguire il riconoscimento del testo. Richiede WebAssembly SIMD, supportato dai browser attuali.',
        captureFailed: 'Impossibile acquisire lo schermo.',
        outside: "L'area selezionata è al di fuori dell'immagine.",
        recognizeFailed: 'Riconoscimento del testo non riuscito.'
      },
      controlRegion: {
        title: 'Calibrazione del mouse',
        description:
          'Utilizza questa impostazione quando il dispositivo controllato usa una risoluzione diversa da 16:9 e il cursore risulta disallineato orizzontalmente o verticalmente.',
        off: 'Disattivata',
        auto: 'Automatica',
        autoWarning:
          "La calibrazione potrebbe non riuscire se l'applicazione utente ha uno sfondo completamente nero.",
        manual: 'Manuale',
        selectedResolution: 'Risoluzione area selezionata',
        unused: 'Non utilizzata',
        originalResolution: 'Risoluzione originale',
        selectResolution: 'Seleziona la risoluzione originale',
        addResolution: 'Aggiungi una risoluzione personalizzata',
        add: 'Aggiungi',
        duplicateResolution: 'Questa risoluzione esiste già.',
        width: 'Larghezza',
        height: 'Altezza',
        apply: 'Calcola e applica',
        invalidResolution: 'Inserisci una risoluzione originale valida quando il video è pronto.',
        select: 'Seleziona area',
        clear: 'Ripristina il rilevamento automatico',
        saveFailed: "Impossibile salvare l'area di input.",
        tooSmall: "L'area selezionata è troppo piccola.",
        previewUnavailable: 'Anteprima non disponibile',
        clearConfirm: 'Ripristinare il rilevamento automatico dei bordi neri?',
        dragHint: "Trascina per selezionare l'area del desktop remoto",
        finish: 'Fine',
        confirm: 'Conferma',
        cancel: 'Annulla'
      },
      auto: 'Automatico',
      autoTips:
        'Potrebbero verificarsi tearing dello schermo o spostamento del mouse a risoluzioni specifiche. Considera di regolare la risoluzione del dispositivo remoto o disabilitare la modalità automatica.',
      fps: 'FPS',
      customizeFps: 'Personalizza',
      quality: 'Qualità',
      qualityLossless: 'Massima',
      qualityHigh: 'Alto',
      qualityMedium: 'Medio',
      qualityLow: 'Basso',
      frameDetect: 'Rilevamento Frame',
      frameDetectTip:
        'Calcola la differenza tra i frame. Interrompe la trasmissione del flusso video quando non vengono rilevate modifiche sullo schermo del dispositivo remoto.',
      resetHdmi: 'Reimposta HDMI',
      mixedH264: {
        title: 'Conflitto del flusso H.264',
        description:
          'I flussi H.264 Direct e H.264 WebRTC sono utilizzati contemporaneamente. Ciò può causare tearing dello schermo o video danneggiato. Utilizzare una sola modalità H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Connessione WebRTC non riuscita',
        description: 'Controlla la connessione di rete o cambia la modalità video.'
      },
      captureStatus: {
        hdmiError: 'Errore schermata HDMI',
        unsupportedResolution: 'La risoluzione attuale non è supportata',
        retrieving: 'Acquisizione schermata...',
        changingResolution: 'Cambio risoluzione...',
        updateFailed: 'Lo schermo non può aggiornarsi al momento',
        videoError: 'Errore di visualizzazione video',
        noHdmi: 'Nessun segnale HDMI rilevato',
        unavailable: 'Lo schermo non può essere visualizzato al momento'
      },
      directConnectionFailed: 'Connessione al flusso video non riuscita'
    },
    keyboard: {
      close: 'Chiudi',
      title: 'Tastiera',
      paste: 'Incolla',
      tips: "Digita il testo sull'host come pressioni di tasti. Scegli il layout di tastiera usato dall'host.",
      placeholder: 'Inserisci testo',
      submit: 'Invia',
      virtual: 'Tastiera',
      readClipboard: 'Leggi dagli Appunti',
      clipboardPermissionDenied:
        "Autorizzazione Appunti negata. Consenti l'accesso agli appunti nel tuo browser.",
      clipboardReadError: 'Impossibile leggere gli appunti',
      mediaKeys: {
        title: 'Tasti multimediali',
        mute: 'Muto',
        volumeDown: 'Abbassa volume',
        volumeUp: 'Alza volume',
        previous: 'Traccia precedente',
        playPause: 'Riproduci o pausa',
        next: 'Traccia successiva',
        stop: 'Stop'
      },
      pasting: {
        layout: "Layout di tastiera dell'host",
        layouts: {
          us: 'Inglese (USA)',
          uk: 'Inglese (Regno Unito)',
          de: 'Tedesco',
          fr: 'Francese',
          es: 'Spagnolo',
          it: 'Italiano',
          ptBr: 'Portoghese (Brasile)',
          se: 'Svedese / finlandese',
          ru: 'Russo',
          ja: 'Giapponese',
          ko: 'Coreano'
        },
        speed: 'Velocità di digitazione',
        speeds: {
          fast: 'Veloce',
          normal: 'Normale',
          slow: 'Lenta'
        },
        estimate: 'Tempo di digitazione: circa {{duration}}',
        untypeable: 'Caratteri che questo layout non può digitare: {{count}}',
        untypeableAt: 'riga {{line}}, colonna {{column}}',
        skipUntypeable: 'Digita il resto',
        shortcut: "{{shortcut}} digita subito gli appunti sull'host.",
        clipboardUnavailable:
          'Il browser consente a una pagina di leggere gli appunti solo tramite HTTPS. Incolla il testo nel riquadro con Ctrl+V.',
        clipboardEmpty: 'Gli appunti non contengono testo.',
        tooLong: 'Il testo è troppo lungo. Il limite è di {{max}} caratteri.',
        inProgress: 'È già in corso la digitazione di un testo incollato.',
        typing: "Digitazione sull'host",
        done: 'Testo digitato',
        canceled: 'Incolla annullato',
        failed: 'Incolla non riuscito',
        cancel: 'Annulla',
        controlBusy: 'Un altro controller sta usando la tastiera.',
        hidError: "Impossibile inviare le pressioni dei tasti all'host."
      },
      shortcut: {
        sendFailed: 'Non inviato: la connessione di input è interrotta',
        title: 'Scorciatoie',
        custom: 'Personalizzato',
        capture: 'Fai clic qui per acquisire il collegamento',
        clear: 'Cancella',
        save: 'Salva',
        captureTips:
          'La cattura dei tasti di sistema (come il tasto Windows) richiede l’autorizzazione a schermo intero.',
        enterFullScreen: 'Attiva/disattiva la modalità a schermo intero.'
      },
      leaderKey: {
        saveFailed: 'Impossibile salvare il tasto leader',
        title: 'Tasto Leader',
        desc: "Ignora le restrizioni del browser e invia collegamenti di sistema direttamente all'host remoto.",
        howToUse: 'Come usare',
        simultaneous: {
          title: 'Modalità simultanea',
          desc1: 'Tieni premuto il tasto Leader, quindi premi la scorciatoia.',
          desc2: 'Intuitivo, ma potrebbe entrare in conflitto con le scorciatoie di sistema.'
        },
        sequential: {
          title: 'Modalità sequenziale',
          desc1:
            'Premi il tasto Leader → premi la scorciatoia in sequenza → premi di nuovo il tasto Leader.',
          desc2: 'Richiede più passaggi, ma evita completamente i conflitti di sistema.'
        },
        enable: 'Abilita tasto Leader',
        tip: 'Quando assegnato come tasto Leader, questo tasto funziona solo come attivatore di scorciatoie e perde il comportamento predefinito.',
        placeholder: 'Premi il tasto Leader',
        shiftRight: 'Shift destro',
        ctrlRight: 'Ctrl destro',
        metaRight: 'Win destro',
        submit: 'Invia',
        recorder: {
          rec: 'REC',
          activate: 'Attiva i tasti',
          input: 'Premi la scorciatoia...'
        }
      }
    },
    mouse: {
      jiggler: 'Movimento automatico del mouse',
      title: 'Mouse',
      cursor: 'Stile cursore',
      default: 'Cursore predefinito',
      pointer: 'Cursore a puntatore',
      cell: 'Cursore a cella',
      text: 'Cursore testo',
      grab: 'Cursore di presa',
      hide: 'Nascondi cursore',
      mode: 'Modalità mouse',
      absolute: 'Modalità assoluta',
      relative: 'Modalità relativa',
      absoluteShort: 'Assoluta',
      relativeShort: 'Relativa',
      touch: 'Modalità touch',
      touchShort: 'Touch',
      absoluteStalled: "L'host sta ignorando il mouse assoluto",
      absoluteStalledDesc:
        "L'host ha smesso di raccogliere i report del mouse assoluto, quindi i movimenti del puntatore vanno persi. La tastiera non è interessata. Ripristinare l'USB di solito risolve; la modalità relativa usa un endpoint diverso.",
      useRelative: 'Passa alla modalità relativa',
      direction: 'Direzione della rotellina',
      scrollUp: 'Come su questo computer',
      scrollDown: 'Invertito (scorrimento naturale)',
      speed: 'Velocità della rotellina',
      fast: 'Veloce',
      slow: 'Lento',
      requestPointer:
        'Usando la modalità relativa. Clicca sul desktop per ottenere il puntatore del mouse.',
      resetHid: 'Reimposta HID',
      hidOnly: {
        switchFailed: 'Impossibile cambiare modalità. Controlla la connessione e riprova.',
        title: 'Modalità solo HID',
        desc: 'Se il mouse e la tastiera smettono di rispondere e il ripristino di HID non aiuta, potrebbe trattarsi di un problema di compatibilità tra IronKVM e il dispositivo. Prova ad abilitare la modalità HID-Only per una migliore compatibilità.',
        tip1: "L'abilitazione della modalità HID-Solo smonterà il disco U virtuale e la rete virtuale",
        tip2: "Nella modalità HID-Only, il montaggio dell'immagine è disabilitato",
        rebuild: 'Cambiare modalità ricostruisce la connessione USB. IronKVM non si riavvia',
        enable: 'Abilita la modalità HID-Solo',
        disable: 'Disabilita la modalità HID-Solo'
      },
      resetHidDone: 'HID USB reimpostato',
      resetHidFailed: 'Reimpostazione HID USB non riuscita'
    },
    image: {
      driveLoaded: 'immagine caricata',
      driveWarning: 'vedi gli avvisi',
      warning: {
        missing:
          "Il file immagine è stato eliminato. L'host legge la vecchia copia finché non la espelli.",
        writable: "Lettura e scrittura: l'host può modificare questa immagine.",
        tooBigForCd: "Troppo grande per l'unità CD ({{size}}, limite {{max}}). Usa il disco.",
        tooSmallForCd: "Troppo piccola per l'unità CD ({{size}}). Usa il disco.",
        empty: 'Il file è vuoto, probabilmente per un caricamento o download non riuscito.'
      },
      delete: 'Elimina',
      inUse: 'In uso. Espellila prima di eliminarla.',
      retry: 'Riprova',
      loadFailed: "Impossibile caricare l'elenco delle immagini",
      readOnlyLocked:
        "Espelli il disco per modificarlo. Si applica quando si inserisce un'immagine.",
      title: 'Immagini',
      loading: 'Caricamento...',
      empty: 'Nessun risultato',
      mountMode: 'Modalità di montaggio',
      mountFailed: 'Montaggio immagine fallito',
      mountDesc:
        "In alcuni sistemi, è necessario espellere il disco virtuale sull'host remoto prima di montare l'immagine.",
      unmountFailed: 'Smontaggio non riuscito',
      unmountDesc:
        "Su alcuni sistemi, è necessario espellere manualmente l'host remoto prima di smontare l'immagine.",
      refresh: "Aggiorna l'elenco delle immagini",
      disk: 'Disco',
      cdrom: 'CD',
      driveEmpty: 'Vuota',
      eject: 'Espelli',
      readOnly: 'Sola lettura',
      readOnlyTip: 'Si applica alla prossima immagine inserita nel disco.',
      noDrives: 'Nessuna unità virtuale. Attiva il disco virtuale nelle Impostazioni.',
      insertFailed: 'Inserimento non riuscito',
      ejectFailed: 'Espulsione non riuscita',
      insertInto: 'Inserisci in {{drive}}. Clicca per cambiare.',
      loadedIn: "Nell'unità {{drive}}",
      attention: 'Attenzione',
      deleteConfirm: 'Sei sicuro di voler eliminare questa immagine?',
      okBtn: 'Sì',
      cancelBtn: 'No',
      deleteFailed: 'Eliminazione non riuscita',
      ventoy: {
        statusNoKernel: 'Non supportato da questo firmware',
        statusNotInstalled: 'Non installato',
        statusReady: 'Pronto',
        statusSelected: 'Immagini selezionate: {{count}}',
        statusInDrive: "Nell'unità disco, {{size}}",
        noKernel:
          "Il kernel di questo firmware non supporta device-mapper, quindi Ventoy non può essere usato finché non si installa un'immagine che lo supporti.",
        installDesc: "Avvia l'host da più immagini su un unico disco, senza copiarle.",
        install: 'Installa',
        installing: 'Download di Ventoy, circa 20 MB. Può richiedere qualche minuto.',
        needsData: "Ventoy richiede un'immagine IronKVM con la partizione /data montata.",
        uninstall: 'Disinstalla',
        uninstallConfirm: 'Rimuovere i file di Ventoy?',
        noImages: 'Nessuna immagine da mettere sul disco Ventoy.',
        onDisk: 'Sul disco Ventoy',
        missing: 'Mancante: {{file}}',
        remove: 'Rimuovi dal disco Ventoy',
        setHint:
          "L'insieme di immagini si può cambiare solo mentre il disco Ventoy non è in nessuna unità.",
        useAsDisk: 'Usa come disco virtuale',
        failed: 'Richiesta Ventoy non riuscita',
        secureBoot:
          "Con Secure Boot attivo, l'host deve registrare una volta la chiave di Ventoy in MokManager. Il file della chiave ENROLL_THIS_KEY_IN_MOKMANAGER.cer si trova nella partizione VTOYEFI.",
        readOnly:
          "L'host vede il disco in sola lettura, quindi la persistenza di Ventoy e ventoy.json sull'unità non funzionano."
      },
      tips: {
        title: 'Come caricare',
        usb1: 'Collega il IronKVM al tuo computer tramite USB.',
        usb2: 'Assicurati che la Virtual Disk sia montata (Impostazioni - Virtual Disk).',
        usb3: 'Apri il disk virtuale sul tuo computer e copia il file immagine nella directory principale del disk.',
        scp1: 'Assicurati che il IronKVM e il tuo computer siano sulla stessa rete locale.',
        scp2: 'Apri un terminale sul tuo computer e usa il comando SCP per caricare il file immagine nella directory /data del IronKVM.',
        scp3: 'Esempio: scp il-tuo-percorso-immagine root@il-tuo-ip-nanokvm:/data',
        tfCard: 'Scheda TF',
        tf1: 'Questo metodo è supportato su sistemi Linux',
        tf2: 'Recupera la scheda TF dal IronKVM (per la versione FULL, smonta prima il case).',
        tf3: 'Inserisci la scheda TF in un lettore di schede e collegala al tuo computer.',
        tf4: 'Copia il file immagine nella directory /data sulla scheda TF.',
        tf5: 'Inserisci la scheda TF nel IronKVM.'
      }
    },
    script: {
      title: 'Script',
      upload: 'Carica',
      run: 'Esegui',
      runBackground: 'Esegui in Background',
      runFailed: 'Esecuzione fallita',
      attention: 'Attenzione',
      delDesc: 'Sei sicuro di voler eliminare questo file?',
      confirm: 'Sì',
      cancel: 'No',
      delete: 'Elimina',
      close: 'Chiudi',
      empty: 'Nessuno script. Carica un file .sh o .py per eseguirlo sulla scheda.',
      loadFailed: 'Impossibile caricare gli script',
      uploaded: 'Script caricato',
      uploadFailed: 'Impossibile caricare lo script',
      started: 'Script avviato in background',
      deleteFailed: 'Impossibile eliminare lo script',
      waitLimit: 'In attesa che lo script finisca, fino a {{minutes}} minuti.',
      timedOut:
        'Lo script è durato più di {{minutes}} minuti e questa pagina ha smesso di attendere. Potrebbe essere ancora in esecuzione sulla scheda.'
    },
    terminal: {
      invalidBaud: 'Questa velocità in baud non è supportata.',
      invalidPort: 'Inserisci un percorso di dispositivo sotto /dev, ad esempio /dev/ttyS1.',
      invalidSettings:
        'Impostazioni della porta seriale non valide. Questa è la shell della scheda.',
      disconnected: 'Disconnesso. Premi Invio per riconnetterti.',
      title: 'Terminale',
      nanokvm: 'Terminale IronKVM',
      serial: 'Terminale Porta Seriale',
      serialPort: 'Porta Seriale',
      serialPortPlaceholder: 'Inserisci la porta seriale',
      baudrate: 'Baud rate',
      parity: 'Parità',
      parityNone: 'Nessuno',
      parityEven: 'Pari',
      parityOdd: 'Dispari',
      flowControl: 'Controllo del flusso',
      flowControlNone: 'Nessuno',
      flowControlSoft: 'Software',
      flowControlHard: 'Hardware',
      dataBits: 'Bit di dati',
      stopBits: 'Bit di stop',
      confirm: 'Ok'
    },
    wol: {
      no: 'No',
      yes: 'Sì',
      deleteConfirm: 'Eliminare questo indirizzo salvato?',
      delete: 'Elimina',
      wake: 'Riattiva',
      rename: 'Rinomina',
      showMac: "Mostra l'indirizzo MAC",
      showName: 'Mostra il nome',
      requestFailed: 'Impossibile raggiungere il dispositivo per inviare il comando',
      deleteFailed: 'Impossibile eliminare',
      renameFailed: 'Impossibile rinominare',
      title: 'Wake-on-LAN',
      sending: 'Invio comando...',
      sent: 'Comando inviato',
      input: 'Inserisci il MAC',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'Caricamento non riuscito',
      uploadSuccess: 'Caricamento completato',
      uploading: 'Caricamento: {{file}}',
      downloadingPercent: 'Download in corso ({{percent}}): {{file}}',
      downloading: 'Download in corso: {{file}}',
      title: 'Scaricatore di immagini',
      input: "Inserisci un'immagine remota URL",
      ok: 'Ok',
      disabled: "La partizione /data è RO, quindi non possiamo scaricare l'immagine",
      uploadbox: 'Rilascia il file qui o fai clic per selezionarlo',
      inputfile: 'Inserisci il file immagine',
      NoISO: 'Nessuna ISO',
      sha256: 'SHA-256 (facoltativo)',
      sha256Placeholder: 'Inserisci un checksum SHA-256 di 64 caratteri',
      invalidSHA256: 'SHA-256 deve essere una stringa esadecimale di 64 caratteri',
      failed: 'Download non riuscito',
      success: 'Download riuscito',
      checksumFailed: 'Download non riuscito: verifica SHA-256 non riuscita',
      cancel: 'Annulla',
      cancelFailed: 'Impossibile annullare il download',
      bootMenu: 'Menu di avvio (netboot.xyz)',
      bootMenuPresent: '{{file}} è già sul dispositivo, con il checksum corretto',
      bootMenuDesc: "Scarica l'ISO di netboot.xyz, con checksum verificato, per il CD virtuale"
    },
    alerts: {
      title: 'Richiede attenzione',
      temperature: {
        warning: "La scheda è a {{celsius}} °C. Controlla che l'aria possa raggiungerla.",
        critical: 'La scheda è a {{celsius}} °C, troppo calda. Dalle aria o spegnila.'
      },
      storage: {
        warning:
          'Solo {{available}} liberi su {{total}} in {{path}}. Le immagini grandi potrebbero non entrare.',
        critical:
          'Solo {{available}} liberi in {{path}}. Caricamenti, download e installazioni di componenti aggiuntivi falliranno. Elimina le immagini che non ti servono.'
      },
      vpn: "{{name}} deve avviarsi al boot ma non è in esecuzione, quindi l'accesso remoto tramite esso non funziona.",
      openVpn: 'Apri impostazioni VPN',
      stream:
        "Lo stream video non è riuscito. Prova un'altra modalità video nel menu Schermo o ricarica la pagina."
    },
    power: {
      resetDesc: "Riavvia subito l'host. Il lavoro non salvato va perso.",
      powerShortDesc: "Accende l'host o chiede al suo sistema operativo di spegnersi (ACPI).",
      powerLongDesc: "Forza lo spegnimento dell'host senza arresto.",
      hddLed: 'LED disco',
      hddActive: 'Attivo',
      hddIdle: 'Inattivo',
      title: 'Accensione',
      showConfirm: 'Conferma',
      showConfirmTip:
        'Chiedi prima di una pressione breve. Il reset e la pressione lunga chiedono sempre.',
      reset: 'Reimposta',
      power: 'Accensione',
      powerShort: 'Accensione (clic breve)',
      powerLong: 'Accensione (clic lungo)',
      resetConfirm: "Procedere con l'operazione di ripristino?",
      powerConfirm: "Procedere con l'operazione di accensione?",
      okBtn: 'Sì',
      cancelBtn: 'No',
      hostOs: "SO dell'host",
      hostOsTip: "Inviati come tasti USB. È l'host a decidere cosa fanno.",
      sleep: 'Sospendi',
      wake: 'Riattiva',
      wakeKey: 'Riattiva con Maiusc',
      powerDown: 'Spegni',
      sleepConfirm: "Sospendere l'host?",
      powerDownConfirm: "Inviare il tasto di spegnimento all'host?",
      wakeTip:
        'Un host sospeso spesso ignora Riattiva dal dispositivo che lo ha sospeso. Riattiva con Maiusc preme un tasto della tastiera, che più host accettano.',
      led: 'LED di accensione',
      ledOn: 'Acceso',
      ledOff: 'Spento',
      ledUnknown: 'Sconosciuto',
      ledConnected: 'LED di accensione collegato',
      ledConnectedTip:
        "Attivalo solo se il connettore del LED di accensione dell'host è cablato alla scheda. Senza, lo stato di alimentazione è sconosciuto.",
      ledConnectedFailed: "Impossibile salvare l'impostazione del LED di accensione",
      powerLongConfirm:
        "Tenere premuto il pulsante di accensione per {{seconds}} s? Toglie l'alimentazione senza spegnimento.",
      done: 'Pulsante premuto',
      failed: 'Pressione del pulsante non riuscita'
    },
    settings: {
      title: 'Impostazioni',
      nav: {
        system: 'Sistema',
        network: 'Rete',
        access: 'Accesso',
        integrations: 'Integrazioni',
        boot: 'Avvio',
        browser: 'Questo browser',
        search: "Cerca un'impostazione",
        noMatch: 'Nessuna impostazione corrisponde',
        locked:
          "Un'operazione è in corso. Le altre pagine e la chiusura non sono disponibili finché non termina.",
        vpnProvider: 'Provider VPN'
      },
      mcp: {
        keyNote:
          'MCP usa una propria chiave API, mostrata qui sotto. Le chiavi della pagina Chiavi API qui non funzionano.',
        title: 'Servizio MCP',
        service: 'Controllo remoto MCP',
        serviceDesc:
          'Consenti ai client MCP attendibili di controllare tastiera e mouse e acquisire schermate',
        securityWarning:
          'Chiunque disponga di questa chiave API può controllare l’host remoto e visualizzarne lo schermo. Usa HTTPS e abilita il servizio solo su reti attendibili.',
        endpoint: 'Endpoint',
        apiKey: 'Chiave API',
        regenerateConfirmTitle: 'Rigenerare la chiave API MCP?',
        regenerateConfirmDesc: 'La chiave attuale smetterà immediatamente di funzionare.',
        enableConfirmTitle: 'Abilitare il controllo MCP esterno?',
        enableConfirmDesc:
          'L’abilitazione di MCP arresterà PicoClaw e chiuderà tutte le sessioni PicoClaw attive.',
        failed: 'Operazione MCP non riuscita',
        copyFailed: 'Copia non riuscita. Copia manualmente.',
        okBtn: 'Conferma',
        cancelBtn: 'Annulla',
        showKey: 'Mostra chiave',
        hideKey: 'Nascondi chiave',
        regenerateKey: 'Rigenera chiave'
      },
      redfish: {
        example: 'Esempio',
        title: 'Redfish',
        service: 'Servizio Redfish',
        serviceDesc:
          "L'API Redfish di DMTF, per il controllo dell'alimentazione, i supporti virtuali e lo stato da strumenti come redfishtool e Ansible. Disattivarla chiude tutte le sessioni Redfish.",
        endpoint: 'Radice del servizio',
        httpsOn: 'La scheda serve HTTPS, necessario per la maggior parte degli strumenti Redfish.',
        httpsOff:
          'La scheda serve HTTP non cifrato. La maggior parte degli strumenti Redfish richiede HTTPS: attivalo in "Impostazioni > Rete".',
        credentials:
          'Redfish accetta gli account del KVM, con autenticazione Basic o una sessione Redfish, e le chiavi API inviate come X-Auth-Token. Le chiavi API si gestiscono nella pagina Chiavi API.',
        powerActions: 'Azioni di alimentazione',
        powerActionsDesc:
          'I tipi di reset offerti ora. On, ForceOff e GracefulShutdown richiedono lo stato di alimentazione, quindi sono offerti solo quando "LED di accensione collegato" è attivo nel menu di accensione.',
        sessions: 'Sessioni',
        noSessions: 'Nessuna sessione Redfish aperta',
        created: 'Creata',
        lastUsed: 'Ultimo utilizzo',
        refresh: 'Aggiorna',
        end: 'Chiudi',
        endConfirmTitle: 'Chiudere questa sessione Redfish?',
        endConfirmDesc:
          'Il suo token smette subito di funzionare. Il client dovrà accedere di nuovo.',
        failed: 'Operazione Redfish non riuscita',
        copyFailed: 'Copia non riuscita. Copia manualmente.',
        okBtn: 'Conferma',
        cancelBtn: 'Annulla'
      },
      ipmi: {
        copyBeforeSave: 'Copia la password ora. Una volta salvata, non può più essere mostrata.',
        noLogin:
          'IPMI è attivo, ma nessun account attivo ha una password IPMI, quindi nessuno può accedere. Impostane una qui sotto.',
        title: 'IPMI',
        warning:
          "L'autenticazione IPMI è debole per progettazione. Chiunque raggiunga la scheda e conosca un nome utente può ottenere un hash della password IPMI di quell'utente e tentare di violarla offline. Usate password generate, attivate IPMI solo su una rete fidata e preferite Redfish su HTTPS quando lo strumento lo supporta.",
        service: 'IPMI su LAN',
        serviceDesc:
          "IPMI 2.0 (RMCP+, ipmitool lanplus) sulla porta UDP 623, per l'alimentazione e lo stato dell'host. IPMI 1.5 e la cipher suite 0 sono rifiutati. Disattivarlo chiude tutte le sessioni IPMI.",
        example: 'Esempio',
        copyFailed: 'Copia non riuscita. Copiare manualmente.',
        ledOn: 'Sono disponibili stato, on, off, soft, cycle e reset.',
        ledOff:
          '"LED di accensione collegato" è disattivato nel menu di alimentazione, quindi lo stato di alimentazione è sconosciuto. Funziona solo "power reset": status, on, off, soft e cycle sono rifiutati.',
        accounts: 'Account',
        accountsDesc:
          'IPMI accede con gli account del KVM, ciascuno con la propria password IPMI, distinta dalla password web. Gli amministratori ottengono ADMINISTRATOR. Gli utenti ottengono USER: possono leggere lo stato di alimentazione con "-L USER" ma non modificarlo.',
        passwordSet: 'Password IPMI impostata',
        passwordNotSet: 'Nessuna password IPMI: accesso IPMI impossibile',
        nameTooLong: 'Il nome supera i 16 caratteri, cosa che IPMI non consente',
        accountDisabled: "L'account è disattivato",
        setPassword: 'Imposta password',
        changePassword: 'Cambia password',
        remove: 'Rimuovi',
        removeConfirmTitle: 'Rimuovere la password IPMI di {{user}}?',
        removeConfirmDesc:
          "L'account non potrà più accedere tramite IPMI e le sue sessioni IPMI terminano.",
        passwordTitle: 'Password IPMI per {{user}}',
        passwordDesc:
          'Da 12 a 20 caratteri ASCII stampabili, diversa dalla password web. IPMI richiede che la scheda conservi la password in una forma rileggibile, quindi usatene una che non sia usata altrove. Copiatela prima di salvare: non verrà più mostrata.',
        passwordPlaceholder: 'Password IPMI',
        generate: 'Genera',
        copy: 'Copia',
        save: 'Salva',
        passwordLength: 'Usare da 12 a 20 caratteri.',
        passwordChars: 'Usare solo caratteri ASCII stampabili.',
        saved: 'Password IPMI salvata',
        failed: 'Operazione IPMI non riuscita',
        okBtn: 'Conferma',
        cancelBtn: 'Annulla'
      },
      ssh: {
        service: 'Server SSH',
        serviceDesc: 'Avvia sshd ora e a ogni avvio',
        failed: 'Impossibile caricare le impostazioni SSH',
        rootDefault: 'root ha ancora la password di fabbrica',
        rootEmpty: 'root non ha una password',
        rootWarning:
          'Chiunque raggiunga la console o SSH può accedere come root. Imposta una password in {{account}} > {{password}}: per il proprietario del dispositivo imposta anche quella di root.',
        connection: 'Connessione',
        command: 'Accedi come root',
        port: 'Porta',
        viaVpn: 'Tramite {{name}}',
        notRunning: 'sshd non è in esecuzione. Attiva il server SSH per collegarti.',
        hostKeys: 'Impronte delle chiavi host',
        hostKeysDesc: 'Confrontale con ciò che ssh mostra alla prima connessione.',
        noHostKeys: 'Nessuna chiave host ancora. sshd le crea al primo avvio.',
        keys: 'Chiavi autorizzate',
        keysDesc:
          'Chiavi pubbliche che possono accedere come root. Sono salvate sulla partizione dati, quindi gli aggiornamenti le conservano.',
        noKeys: 'Ancora nessuna chiave autorizzata.',
        noComment: 'nessun commento',
        addPlaceholder:
          'Incolla una chiave pubblica, ad esempio il contenuto di ~/.ssh/id_ed25519.pub',
        add: 'Aggiungi chiave',
        added: 'Chiave aggiunta',
        removed: 'Chiave rimossa',
        deleteConfirm: 'Rimuovere questa chiave?',
        deleteConfirmDesc: 'Non potrà più accedere. Le sessioni aperte restano aperte.',
        invalidKey: 'Questa non è una chiave pubblica. Incolla una sola riga da un file .pub.',
        keyOptions: 'Qui non sono accettate chiavi con opzioni come command= o from=.',
        duplicateKey: 'Questa chiave è già autorizzata.',
        lastKey:
          "L'ultima chiave non può essere rimossa mentre l'accesso solo con chiavi è attivo.",
        keysOnly: 'Solo chiavi',
        keysOnlyDesc:
          "Disattiva l'accesso con password e keyboard-interactive. Le sessioni aperte restano aperte.",
        keysOnlyNeedsKey:
          'Aggiungi prima una chiave autorizzata, altrimenti nessuno potrebbe accedere.',
        keysOnlyOn: 'Accesso con password disattivato',
        keysOnlyOff: 'Accesso con password attivato',
        notHonoured:
          "Lo sshd di questa immagine non legge questa impostazione, quindi l'accesso con password resta attivo.",
        reloadFailed:
          'Salvato, ma non è stato possibile ricaricare sshd. Verrà applicato al prossimo avvio di sshd.',
        notApplied:
          "sshd accetta ancora le password. Spegni e riaccendi il server SSH per applicare l'impostazione."
      },
      vnc: {
        address: 'Indirizzo',
        certHint:
          "VeNCrypt X509Plain usa il certificato autofirmato del dispositivo, quindi il client avvisa alla prima connessione. Accettalo, oppure salva il certificato dall'indirizzo HTTPS di questa pagina e passalo a TigerVNC con -X509CA=<file>.",
        title: 'VNC',
        service: 'Server VNC',
        serviceDesc:
          "Consente a un client VNC, come TigerVNC o Remmina, di visualizzare e controllare l'host. Il client deve supportare la codifica Tight. Una sessione alla volta.",
        credentials:
          'Accedi con un account KVM. La connessione è cifrata con il certificato TLS della scheda (VeNCrypt X509Plain).',
        port: 'Porta',
        portDesc: 'La porta TCP su cui il server è in ascolto.',
        maxFps: 'Limite di fotogrammi',
        maxFpsDesc: 'Il numero massimo di fotogrammi al secondo inviati a un client.',
        vncAuth: 'Autenticazione VNC semplice',
        vncAuthDesc:
          'Per i client senza VeNCrypt. Verifica una password VNC separata invece di un account.',
        vncAuthWarning:
          "L'autenticazione VNC semplice non cifra la connessione. Chiunque sul percorso di rete può vedere lo schermo e i tasti premuti. Usala solo su una rete fidata.",
        password: 'Password VNC',
        passwordSet: 'È impostata una password. Digitane una nuova per cambiarla.',
        passwordInvalid: 'La password VNC deve essere di 6-8 caratteri.',
        save: 'Salva',
        saved: 'Impostazioni salvate',
        state: 'Stato',
        listening: 'In ascolto sulla porta {{port}}',
        notListening: 'Non in ascolto',
        noSession: 'Nessuna sessione aperta',
        client: 'Client',
        user: 'Utente',
        method: 'Autenticazione',
        methodVencrypt: 'Account su TLS',
        methodVnc: 'Password VNC',
        since: 'Connesso dal',
        resolution: 'Risoluzione',
        framesSent: 'Fotogrammi inviati',
        lastError: "L'ultima sessione è terminata: {{error}}",
        refresh: 'Aggiorna',
        disconnect: 'Disconnetti',
        disconnectConfirmTitle: 'Terminare la sessione VNC?',
        disconnectConfirmDesc:
          'Il client viene disconnesso subito, e ogni tasto e pulsante che tiene premuto viene rilasciato.',
        failed: 'Operazione VNC non riuscita',
        okBtn: 'Conferma',
        cancelBtn: 'Annulla'
      },
      watchdog: {
        title: 'Watchdog',
        service: "Watchdog dell'host",
        serviceDesc:
          "Se l'host dovrebbe essere acceso e la sua immagine non cambia, o manca il segnale HDMI, per il tempo di attesa, la scheda preme reset o spegne e riaccende l'host.",
        stillWarning:
          "Un host il cui schermo va in sospensione, o la cui immagine resta ferma mentre lavora, sembra bloccato. Disattivate la sospensione dello schermo sull'host o impostate un indirizzo di ping.",
        ledHint:
          '"LED di accensione collegato" è disattivato nel menu di alimentazione. Il watchdog non vede quando l\'host è spento, quindi lo considera sempre acceso.',
        timeout: 'Tempo di attesa',
        timeoutDesc:
          "Per quanto tempo l'host può non dare segni di vita prima che il watchdog intervenga.",
        action: 'Azione',
        actionDesc:
          'Il ciclo di alimentazione tiene premuto il pulsante di accensione per 5 secondi, poi lo preme di nuovo.',
        actionReset: 'Reimposta',
        actionPower: 'Ciclo di alimentazione',
        cooldown: 'Pausa',
        cooldownDesc: 'Il tempo minimo tra due azioni.',
        maxPerHour: 'Azioni per ora',
        maxPerHourDesc: 'Il numero massimo di azioni in un’ora.',
        pingHost: 'Indirizzo di ping',
        pingHostDesc:
          "L'indirizzo IP dell'host. Una risposta conta come segno di vita. Lasciate vuoto per non fare ping.",
        pingHostInvalid: 'Inserite un indirizzo IPv4 o IPv6.',
        minutes: 'min',
        save: 'Salva',
        saved: 'Salvato',
        state: 'Rilevatore',
        status: {
          off: 'Disattivato',
          watching: 'In osservazione',
          hostOff: 'Host spento',
          captureOff: 'Acquisizione HDMI disattivata',
          cooldown: 'In pausa',
          capped: 'Limite orario raggiunto',
          acting: 'In azione'
        },
        signal: 'Segnale HDMI',
        yes: 'Sì',
        no: 'No',
        led: 'LED di accensione',
        on: 'Acceso',
        off: 'Spento',
        ledNotConnected: 'Non collegato',
        ping: 'Ping',
        pingNotSet: 'Non impostato',
        pingReply: 'Risponde',
        pingNoReply: 'Nessuna risposta',
        lastChange: "Ultimo cambio d'immagine",
        never: 'Mai',
        actsIn: 'Interviene tra',
        actionsLastHour: "Azioni nell'ultima ora",
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Registro',
        noLog: 'Il watchdog non è ancora intervenuto.',
        refresh: 'Aggiorna',
        reasonFrozen: "L'immagine non è cambiata",
        reasonNoSignal: 'Nessun segnale HDMI',
        stuckFor: 'nessun segno di vita per {{duration}}',
        pressFailed: 'La pressione non è riuscita: {{error}}',
        noScreenshot: 'Nessuno screenshot',
        failed: 'Operazione del watchdog non riuscita',
        powerNeedsLed:
          'Il ciclo di alimentazione richiede "LED di alimentazione collegato" nel menu di alimentazione.',
        noLedConfirmTitle: 'Attivare il watchdog senza il LED di alimentazione?',
        noLedConfirmDesc:
          "La scheda non vede quando l'host è spento, quindi lo considera sempre acceso. Se spegni l'host, il watchdog preme reset allo scadere del timeout. Collega il LED di alimentazione per evitarlo.",
        noLedConfirmOk: 'Attiva',
        cancel: 'Annulla'
      },
      netboot: {
        title: 'Avvio di rete',
        description:
          "Avviare l'host dalla rete: iPXE e un menu delle immagini sul KVM tramite il collegamento di rete USB, oppure netboot.xyz tramite proxy DHCP sulla LAN.",
        addon: 'dnsmasq e file di avvio',
        addonDesc:
          'Installati su /data: dnsmasq da Alpine, iPXE e netboot.xyz dalle loro release, ciascuno verificato con il proprio checksum.',
        install: 'Installa',
        installing: 'Installazione in corso. Può richiedere alcuni minuti.',
        uninstall: 'Disinstalla',
        uninstallConfirm: "Disattivare l'avvio di rete e rimuovere dnsmasq e i file di avvio?",
        needsData: "L'avvio di rete richiede un'immagine IronKVM con la partizione /data montata.",
        usb: 'Sul collegamento di rete USB',
        usbDesc:
          "Finché il collegamento di rete USB è attivo, dnsmasq lo serve al posto di udhcpd. L'host riceve il suo unico indirizzo senza router né server DNS, iPXE per la sua architettura e un menu delle immagini ISO sul KVM.",
        linkOff: 'Il collegamento di rete USB è disattivato. Attivalo in Dispositivo, Rete USB.',
        menuUrl: 'Menu',
        leases: "Lease dell'host",
        noLeases: 'Ancora nessuno',
        netbootxyzNote:
          "netboot.xyz nel menu si carica da internet, che il collegamento USB non raggiunge. L'host ha bisogno di internet su un'altra porta di rete.",
        lan: 'Proxy DHCP sulla LAN',
        lanDesc:
          'Risponde ai client PXE sulla LAN con netboot.xyz, che poi carica il suo menu da internet. Non assegna mai indirizzi e non serve le immagini sul KVM.',
        lanWarning:
          "netboot.xyz viene offerto a ogni client PXE di questa LAN, non solo all'host. Attivalo solo su una rete che controlli.",
        lanConfirm: 'Attivare il proxy DHCP sulla LAN?',
        lanInterface: 'LAN',
        running: 'In esecuzione',
        stopped: 'Non in esecuzione',
        images: 'Immagini nel menu',
        noImages: 'Nessuna immagine ISO nella directory delle immagini.',
        boots: 'Avvii recenti',
        noBoots: "L'host non ha ancora scaricato nulla.",
        log: 'Log di dnsmasq',
        refresh: 'Aggiorna',
        okBtn: 'Conferma',
        cancelBtn: 'Annulla',
        failed: 'Operazione di avvio di rete non riuscita'
      },
      about: {
        title: 'Informazioni su IronKVM',
        information: 'Informazioni',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versione Applicazione',
        applicationTip: 'Versione dell’applicazione web IronKVM',
        image: 'Versione Immagine',
        imageTip: "Immagine della scheda IronKVM e l'immagine di sistema NanoKVM su cui si basa",
        kernel: 'Versione del kernel',
        kernelTip: 'Versione del kernel Linux attualmente in esecuzione',
        deviceKey: 'Chiave Dispositivo',
        videoMemory: 'Memoria video',
        videoMemoryTip:
          'Memoria riservata alla cattura video. Non è condivisa con il resto del sistema.',
        videoMemoryGenerations_one:
          '{{count}} sessione precedente di IronKVM sta trattenendo memoria video',
        videoMemoryGenerations_other:
          '{{count}} sessioni precedenti di IronKVM stanno trattenendo memoria video',
        videoMemoryReboot: 'Riavvia per recuperarla.',
        community: 'Comunità',
        hostname: 'Nome host',
        hostnameUpdated: 'Nome host aggiornato. Riavviare per applicare.',
        ipType: {
          Wired: 'Cablato',
          Wireless: 'Senza fili',
          Other: 'Altro'
        },
        hostnameInvalid:
          "Usa lettere, cifre e trattini, fino a 63 per parte separata da punti. Nessun trattino all'inizio o alla fine di una parte.",
        hostnameFailed: 'Impossibile modificare il nome host',
        editHostname: 'Modifica nome host',
        docs: 'Documentazione',
        hardware: 'Hardware',
        hardwareFaq: 'FAQ hardware',
        disclaimer:
          'IronKVM: firmware comunitario rafforzato per il Sipeed NanoKVM. Non affiliato a Sipeed.',
        basedOn: 'basato su NanoKVM {{version}}'
      },
      preferences: {
        title: 'Preferenze'
      },
      performance: {
        title: 'Prestazioni'
      },
      appearance: {
        thisBrowser: 'Questo browser',
        thisBrowserDesc:
          'Salvato solo in questo browser. Gli altri browser hanno le proprie impostazioni.',
        deviceWide: 'Dispositivo',
        deviceWideDesc: 'Salvato sul dispositivo. Vale per chiunque lo apra.',
        language: 'Lingua',
        languageDesc: "Seleziona la lingua per l'interfaccia",
        webTitle: 'Titolo web',
        webTitleDesc: 'Personalizza il titolo della pagina web',
        menuBar: {
          title: 'Barra dei menu',
          mode: 'Modalità di visualizzazione',
          modeDesc: 'Visualizza la barra dei menu sullo schermo',
          modeOff: 'Spento',
          modeAuto: 'Nascondi automaticamente',
          modeAlways: 'Sempre visibile',
          keyboardLedStatus: 'Indicatori di blocco della tastiera',
          keyboardLedStatusDesc:
            'Mostra lo stato di Bloc Num, Bloc Maiusc e Bloc Scorr del computer remoto',
          icons: 'Icone dei sottomenu',
          iconsDesc: 'Visualizza le icone dei sottomenu nella barra dei menu'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Stato dei blocchi della tastiera remota',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Bloc Num',
        numLockShort: 'Num',
        capsLock: 'Bloc Maiusc',
        capsLockShort: 'Mai',
        scrollLock: 'Bloc Scorr',
        scrollLockShort: 'Scorr',
        on: 'Attivo',
        off: 'Disattivo',
        unknown: 'Sconosciuto'
      },
      device: {
        title: 'Dispositivo',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'Luminosità OLED',
          brightnessDescription: 'Un livello più basso allunga la vita del display',
          brightnessLevels: {
            '64': 'Minima',
            '96': 'Bassa',
            '128': 'Media',
            '160': 'Alta',
            '207': 'Predefinita',
            '255': 'Massima'
          },
          0: 'Mai',
          15: '15 sec',
          30: '30 sec',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 ora'
        },
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Pannello frontale'
        },
        hidModeDesc:
          "Prova la modalità solo HID se l'host non accetta tastiera e mouse. Disattiva le unità virtuali e la rete.",
        resetHidDesc: "Ricollega tastiera e mouse all'host. Usalo se l'input smette di funzionare.",
        cpuFreq: {
          title: 'Frequenza CPU',
          description: 'Imposta la frequenza della CPU applicata al prossimo avvio',
          tip: 'La CPU si avvia a 850 MHz ed è specificata per 1000 MHz. Un nuovo valore si applica al prossimo avvio, non mentre il sistema è in funzione. 1000 MHz rientra nelle specifiche; la temperatura resta ben entro i limiti con entrambe le impostazioni.',
          running: 'In uso: {{mhz}} MHz',
          rebootToApply: 'riavvia per applicare',
          rebootConfirm: 'Riavviare ora per applicare {{mhz}} MHz?'
        },
        swap: {
          title: 'Scambia',
          disable: 'Disabilita',
          description: 'Imposta la dimensione del file di scambio',
          tip: 'Abilitare questa funzione potrebbe ridurre la durata utile della tua scheda SD!'
        },
        zram: {
          title: 'Swap compresso (zram)',
          description: 'Swap nella RAM compressa, invece che sulla scheda SD',
          tip: "zram tiene lo swap lontano dalla scheda SD, quindi non la usura. Non c'è uno swap su disco dietro: se zram si riempie, il kernel termina un processo invece di paginare lentamente. Il limite di memoria stabilisce quanta RAM può occupare zram.",
          unavailable: 'I moduli del kernel non sono installati su questo dispositivo',
          inactive: 'Abilitato, ma il dispositivo non è stato avviato',
          active: 'Attivo - {{used}} di {{total}}, {{ratio}}x',
          off: 'Disattivato',
          detail: {
            algorithm: 'Algoritmo: {{algorithm}}',
            memory: 'Memoria usata: {{used}} di {{limit}}',
            memoryNoLimit: 'Memoria usata: {{used}}, nessun limite impostato',
            counters:
              "Pagine scambiate: in entrata {{in}}, in uscita {{out}} (tutti i dispositivi di swap, dall'avvio)"
          }
        },
        mouseJiggler: {
          title: 'Muovi il mouse',
          description: "Impedisce la sospensione dell'host remoto",
          disable: 'Disabilita',
          absolute: 'Modalità assoluta',
          relative: 'Modalità relativa'
        },
        mdns: {
          description: 'Abilita il servizio di rilevamento mDNS',
          tip: 'Spegnerlo se non è necessario'
        },
        hdmi: {
          description: 'Abilita HDMI/monitora uscita',
          idleTimeoutTitle: 'Timeout cattura inattiva',
          idleTimeoutDescription:
            'Interrompi la cattura HDMI dopo che non ci sono visualizzatori attivi per',
          minutes: 'min'
        },
        hidOnly: 'HID-Solo modalità',
        hidOnlyDesc:
          'Smette di emulare i dispositivi virtuali, mantenendo solo il controllo di base HID',
        disk: 'Disco virtuale',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Rete virtuale',
        networkDesc: 'Monta la scheda di rete virtuale sull’host remoto',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Host:',
          description:
            "Un collegamento di rete privato con l'host remoto tramite il cavo USB. L'host riceve un indirizzo senza gateway né DNS, quindi non può raggiungere la tua LAN tramite IronKVM.",
          off: 'Disattivato',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (per host senza NCM)',
          rndis: 'RNDIS (non più offerto)',
          rndisNote: 'Questo collegamento usa RNDIS, che non è più offerto. Scegli NCM o ECM.',
          subnet: 'Sottorete',
          subnetDesc:
            "Una rete IPv4 privata, da /24 a /30. IronKVM prende il primo indirizzo, l'host il secondo.",
          invalidSubnet: 'Inserisci una sottorete come 172.31.255.0/30.',
          apply: 'Applica',
          confirm: 'Ricollegare il dispositivo USB?',
          reenumerate:
            "Applicando si ricostruisce la connessione USB. L'host perde tastiera, mouse e disco virtuale per alcuni secondi."
        },
        audio: 'Altoparlante virtuale',
        audioDesc:
          "Presenta una scheda audio USB all'host remoto, così puoi sentirlo. L'host deve selezionarla come dispositivo di uscita. Cambiare questa opzione ricostruisce la connessione USB.",
        audioNote:
          "L'audio è disponibile in entrambe le modalità H.264 (WebRTC e Direct), non in MJPEG",
        console: 'Console seriale',
        consoleDesc:
          "Presenta una porta seriale USB all'host remoto, per accedere a questo IronKVM quando la rete non è raggiungibile",
        consoleTip:
          "Chiunque controlli l'host remoto ottiene un prompt di accesso a questo IronKVM. Imposta una password complessa prima dell'abilitazione (Account - Modifica password).",
        endpoints: {
          title: 'Endpoint USB',
          used: '{{used}} di {{total}} in uso',
          cost: 'ne usa {{cost}}',
          needs: 'ne richiede {{cost}}',
          full: "Endpoint USB insufficienti. Disattiva prima qualcos'altro.",
          inactive:
            'Attivo, ma non in funzione: il controller USB ha esaurito gli endpoint. Disattiva un altro dispositivo e questo si avvia subito.',
          explain:
            'Il controller USB ha un numero fisso di endpoint in ingresso, e questo li conta. Se sono abilitati più dispositivi di quanti ne entrino, tastiera e mouse vengono mantenuti e gli altri disattivati.',
          error: 'Impossibile raggiungere il dispositivo. Riprova.',
          fitTogether: 'Stanno insieme: {{sets}}'
        },
        reboot: 'Riavvia',
        rebootDesc: 'Sei sicuro di voler riavviare IronKVM?',
        okBtn: 'Sì',
        cancelBtn: 'No',
        rebootFailed: 'Riavvio non riuscito'
      },
      network: {
        title: 'Rete',
        wifi: {
          disconnectBtn: 'Disconnetti',
          disconnectWarning:
            'Se raggiungi IronKVM tramite questa rete Wi-Fi, questa pagina perderà la connessione.',
          disconnected: 'Wi-Fi disconnesso',
          title: 'Wi-Fi',
          description: 'Configura Wi-Fi',
          apMode: 'La modalità AP è attiva, connettiti al Wi-Fi scansionando il codice QR',
          connect: 'Connetti Wi-Fi',
          connectDesc1: 'Inserisci SSID e password della rete',
          connectDesc2: 'Inserisci la password per unirti a questa rete',
          disconnect: 'Vuoi davvero disconnettere la rete?',
          failed: 'Connessione non riuscita, riprova.',
          ssid: 'Nome',
          password: 'Password',
          joinBtn: 'Connetti',
          confirmBtn: 'OK',
          cancelBtn: 'Annulla'
        },
        tls: {
          description: 'Abilita protocollo HTTPS',
          tip: "Attenzione: l'uso di HTTPS può aumentare la latenza, soprattutto in modalità video MJPEG.",
          restarting: 'Riavvio del server del dispositivo, ci vorranno circa due minuti...',
          waiting: 'In attesa che il dispositivo risponda di nuovo...',
          waitingHttp: 'Ritorno a http. Ricarica questa pagina se non si apre da sola.',
          failed: "Impossibile modificare l'impostazione HTTPS",
          enableConfirm: 'Attivare HTTPS?',
          disableConfirm: 'Disattivare HTTPS?',
          confirmDesc:
            'Questo ti disconnette e riavvia il server del dispositivo, operazione che richiede circa due minuti. Poi la pagina apre {{url}}.',
          confirmOk: 'Continua',
          confirmCancel: 'Annulla'
        },
        ethernet: {
          title: 'Indirizzo IP',
          description: 'Configura come IronKVM ottiene il suo indirizzo sulla rete cablata',
          dhcp: 'DHCP',
          manual: 'Manuale',
          networkDetails: 'Dettagli di rete',
          interface: 'Interfaccia',
          ipAddress: 'Indirizzo IP',
          subnetMask: 'Maschera di sottorete',
          router: 'Router',
          save: 'Applica',
          invalidAddress: 'Inserisci un indirizzo IP valido',
          invalidMask: 'Inserisci una maschera di sottorete valida, ad esempio 255.255.255.0 o 24',
          invalidRouter: 'Inserisci un indirizzo del router valido',
          addressRequired: 'È richiesto un indirizzo IP',
          maskRequired: 'È richiesta una maschera di sottorete',
          applyTitle: "Cambiare l'indirizzo di IronKVM?",
          applyWarning:
            "La connessione a questa pagina andrà persa. IronKVM applica il nuovo indirizzo e attende {{seconds}} secondi che tu lo raggiunga a quell'indirizzo. Raggiungerlo mantiene la modifica. Se non lo raggiunge nulla, IronKVM ripristina le impostazioni precedenti.",
          applyConfirm: 'Applica',
          applyCancel: 'Annulla',
          applyFailed: "Impossibile applicare l'indirizzo",
          trialTitle: 'In attesa di conferma',
          trialDhcp: 'IronKVM sta chiedendo un indirizzo al DHCP.',
          trialStatic: 'IronKVM ora si trova a {{address}}.',
          trialInstruction:
            'Apri IronKVM al suo nuovo indirizzo e accedi se lo chiede. Raggiungerlo lì mantiene la modifica. Se nulla raggiunge IronKVM entro {{seconds}} secondi, ripristina le impostazioni precedenti.',
          trialOpen: 'Apri il nuovo indirizzo',
          trialKeep: 'Mantieni queste impostazioni',
          trialKept: 'Il nuovo indirizzo è salvato',
          trialKeepFailed: 'Impossibile mantenere le impostazioni',
          trialGone: 'La modifica è già stata ripristinata. Riprova.',
          unsaved: 'Modifiche non salvate'
        },
        dns: {
          title: 'DNS',
          description: 'Configura i server DNS per IronKVM',
          mode: 'Modalità',
          dhcp: 'DHCP',
          manual: 'Manuale',
          add: 'Aggiungi DNS',
          save: 'Salva',
          invalid: 'Inserisci un indirizzo IP valido',
          noDhcp: 'Nessun DNS DHCP è attualmente disponibile',
          saved: 'Impostazioni DNS salvate',
          saveFailed: 'Impossibile salvare le impostazioni DNS',
          unsaved: 'Modifiche non salvate',
          maxServers: 'Sono consentiti al massimo {{count}} server DNS',
          dnsServers: 'Server DNS',
          dhcpServersDescription: 'I server DNS vengono ottenuti automaticamente da DHCP',
          manualServersDescription: 'I server DNS possono essere modificati manualmente',
          networkDetails: 'Dettagli rete',
          interface: 'Interfaccia',
          ipAddress: 'Indirizzo IP',
          subnetMask: 'Subnet mask',
          router: 'Router',
          none: 'Nessuno'
        }
      },
      vpn: {
        connect: 'Connetti',
        connectDesc: 'Entra nella rete {{name}}. Spento disconnette senza fermare il servizio.',
        kvmUrl: 'Indirizzo del KVM',
        moreTip: 'Altre azioni',
        restartTip: 'Riavvia',
        stopTip: 'Ferma',
        updateTip: 'Aggiorna a {{version}}',
        loading: 'Caricamento...',
        okBtn: 'Sì',
        cancelBtn: 'No',
        restart: 'Riavviare {{name}}?',
        stop: 'Arrestare {{name}}?',
        stopDesc:
          "Il demone si arresta ora. Avvio all'accensione è un interruttore separato e resta com'è.",
        update: 'Aggiornare {{name}} alla versione {{version}}?',
        updateDesc: "Il demone si riavvia se è in esecuzione. L'accesso viene mantenuto.",
        notInstall: '{{name}} non è installato.',
        install: 'Installa',
        installing: 'Installazione',
        installFailed: 'Installazione non riuscita',
        retry: 'Riprova',
        notRunning: '{{name}} non è in esecuzione. Avvialo per continuare.',
        run: 'Avvia',
        boot: "Avvio all'accensione",
        bootDesc: "Avvia {{name}} all'accensione del KVM.",
        control: 'Server di controllo',
        connected: 'Connesso',
        disconnected: 'Non connesso',
        deviceName: 'Nome dispositivo',
        deviceIP: 'IP dispositivo',
        account: 'Account',
        version: 'Versione',
        uptime: 'Tempo di attività',
        peers: 'Peer',
        noPeers: 'Ancora nessun peer.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Memoria',
        daemonRss: 'Demone',
        group: 'Gruppo componenti aggiuntivi',
        high: 'rallentato oltre {{size}}',
        max: 'terminato dal kernel oltre {{size}}',
        noGroup: 'Nessun gruppo di memoria per componenti aggiuntivi su questa scheda.',
        uninstall: 'Disinstalla {{name}}',
        uninstallDesc: "Sei sicuro di voler disinstallare {{name}}? L'accesso resta sulla scheda.",
        blocked:
          "{{other}} è in esecuzione o si avvia all'accensione. Può funzionare una sola VPN alla volta: prima arresta {{other}} e disattivane l'avvio all'accensione.",
        swap: {
          title: 'Memoria di swap',
          tip: 'Se il demone resta a corto di memoria, prova ad attivare lo swap. Si imposta in "Impostazioni > Prestazioni".'
        },
        copy: 'Copia',
        copied: 'Link copiato',
        copyFailed: 'Impossibile copiare il link. Selezionalo e copialo a mano.',
        open: 'Apri',
        checkAgain: 'Controlla di nuovo',
        notSignedIn:
          "Accesso non ancora eseguito. Completa l'accesso dal link, poi controlla di nuovo.",
        checkFailed: 'Impossibile verificare lo stato di accesso',
        loginWaiting: "Questa pagina controlla ogni pochi secondi e prosegue dopo l'accesso.",
        uninstallFailed: 'Disinstallazione non riuscita',
        loginFailed: 'Accesso non riuscito'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Scarica il',
        package: 'pacchetto di installazione',
        unzip: 'e decomprimilo',
        notLogin:
          'Il dispositivo non è ancora stato associato. Effettua il login e associa questo dispositivo al tuo account.',
        urlPeriod: 'Questo URL è valido per 10 minuti',
        login: 'Accedi',
        logout: 'Disconnetti',
        logoutDesc: 'Sei sicuro di voler uscire?',
        manualIntro: 'Oppure installalo a mano via SSH:',
        copyBinaries: "Copia tailscale e tailscaled in {{dir}} sull'IronKVM",
        linksFile: 'Nella stessa directory, crea un file chiamato links con queste due righe:',
        rebootRefresh: "Riavvia l'IronKVM, poi aggiorna questa pagina"
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Questo dispositivo non si è ancora unito a una rete NetBird. Unisciti con una chiave di configurazione o accedi con SSO.',
        setupKey: 'Chiave di configurazione',
        setupKeyPlaceholder: 'Incolla una chiave di configurazione dalla dashboard di NetBird',
        join: 'Unisciti',
        or: 'oppure',
        sso: 'Accedi con SSO',
        urlPeriod: 'Questo URL è valido per 10 minuti',
        logout: 'Rimuovi registrazione',
        logoutDesc:
          'Rimuovere la registrazione elimina questo peer dal tuo account NetBird e ne cancella qui la configurazione. Per unirti di nuovo serve una chiave di configurazione o un accesso SSO, e il peer potrebbe ricevere un nuovo IP. Continuare?',
        joinFailed: 'Impossibile unirsi alla rete'
      },
      update: {
        title: 'Controlla Aggiornamenti',
        queryFailed: 'Impossibile ottenere la versione',
        updateFailed: 'Aggiornamento fallito. Riprova.',
        isLatest: 'Hai già la versione più recente.',
        available: 'Un aggiornamento è disponibile. Sei sicuro di voler aggiornare?',
        updating: 'Aggiornamento avviato. Attendere prego...',
        confirm: 'Conferma',
        cancel: 'Annulla',
        preview: 'Anteprima aggiornamenti',
        previewDesc: "Ottieni l'accesso anticipato a nuove funzionalità e miglioramenti",
        previewTip:
          'Tieni presente che le versioni di anteprima possono contenere bug o funzionalità incomplete!',
        customServer: {
          title: 'Server di aggiornamento personalizzato',
          desc: 'Cerca e scarica gli aggiornamenti online da un server specificato',
          invalidUrl:
            'Inserisci una directory server HTTP o HTTPS valida, senza parametri di query, frammenti o latest.json.',
          loadFailed: 'Impossibile caricare la configurazione del server di aggiornamento.',
          saveFailed: 'Impossibile salvare la configurazione del server di aggiornamento.',
          saved: 'Configurazione del server di aggiornamento salvata.',
          save: 'Salva',
          confirmTitle: 'Utilizzare un server di aggiornamento personalizzato?',
          confirmDesc:
            'SHA-512 verifica soltanto che il pacchetto corrisponda al manifesto fornito da questo server. Non garantisce che il pacchetto sia una versione ufficiale di IronKVM. Un server difettoso o dannoso può rendere inutilizzabile il dispositivo, causare la perdita di dati o compromettere il sistema.',
          confirm: 'Utilizza comunque',
          useSipeed: 'Usa il server ufficiale di Sipeed',
          previewDisabled:
            'Gli aggiornamenti in anteprima non sono disponibili quando è attivo un server di aggiornamento personalizzato.'
        },
        offline: {
          chooseFile: 'Scegli file',
          installing: 'Caricamento completato. Installazione...',
          noFile: 'Nessun file scelto',
          title: 'Aggiornamenti offline',
          desc: 'Aggiornamento tramite pacchetto di installazione locale',
          upload: 'Carica',
          checksumPlaceholder: 'Checksum SHA-256 (facoltativo)',
          invalidChecksum: 'Il checksum SHA-256 deve contenere 64 caratteri esadecimali.',
          checksumMismatch:
            'La verifica SHA-256 non è riuscita. Il pacchetto potrebbe essere danneggiato.',
          invalidName: 'Formato nome file non valido. Si prega di scaricare dalle versioni GitHub.',
          updateFailed: 'Aggiornamento fallito. Riprova.'
        },
        updateTo: 'Aggiorna a {{version}}',
        updateConfirmDesc:
          "Il dispositivo installa l'aggiornamento e riavvia il server. La pagina si ricarica quando il server torna disponibile.",
        releaseNotes: 'Note di rilascio'
      },
      account: {
        title: 'Account',
        webAccount: 'Nome account web',
        role: 'Ruolo',
        roles: { admin: 'Amministratore', user: 'Utente' },
        password: 'Password',
        updateBtn: 'Update',
        logoutBtn: 'Esci',
        logoutDesc: 'Sei sicuro di voler uscire?',
        okBtn: 'Sì',
        cancelBtn: 'No',
        users: {
          title: 'Utenti',
          create: 'Crea utente',
          enabled: 'Abilitato',
          disabled: 'Disabilitato',
          deviceOwner: 'Proprietario del dispositivo',
          resetPassword: 'Reimposta password',
          delete: 'Elimina',
          deleteConfirm: 'Eliminare questo utente e revocare tutte le sue sessioni?',
          created: 'Utente creato',
          deleted: 'Utente eliminato',
          passwordUpdated: 'Password aggiornata',
          loadFailed: 'Impossibile caricare gli utenti',
          saveFailed: "Impossibile salvare l'utente",
          deleteFailed: "Impossibile eliminare l'utente"
        }
      },
      apiKeys: {
        mcpNote: 'Queste chiavi non valgono per MCP, che ha una propria chiave nella pagina MCP.',
        metricsUrl: 'URL delle metriche',
        monitoring: 'Monitoraggio',
        monitoringDesc:
          'Prometheus legge le metriche con una chiave API di questa pagina, inviata come token Bearer. Qualsiasi ruolo può leggerle.',
        scrapeConfig: 'Configurazione scrape di Prometheus',
        title: 'Chiavi API',
        description:
          "Una chiave agisce come il suo proprietario, con il ruolo di quell'utente. Inviala come Authorization: Bearer <key> per le metriche e l'API, o come X-Auth-Token per Redfish.",
        name: 'Nome',
        namePlaceholder: 'A cosa serve la chiave, ad esempio prometheus',
        nameRequired: 'Dai un nome alla chiave',
        nameTooLong: 'Il nome può contenere al massimo 64 caratteri',
        unnamed: '(senza nome)',
        create: 'Crea chiave',
        created: 'Creata',
        owner: 'Proprietario',
        empty: 'Nessuna chiave API',
        newKeyTitle: 'La tua nuova chiave API',
        newKeyWarning:
          "Copia subito la chiave. Non viene memorizzata e non potrà essere mostrata di nuovo. Se la perdi, revocala e creane un'altra.",
        copy: 'Copia',
        copied: 'Copiata',
        copyFailed: 'Copia non riuscita. Copia manualmente.',
        done: 'Fatto',
        revoke: 'Revoca',
        revokeConfirmTitle: 'Revocare questa chiave API?',
        revokeConfirmDesc: 'Tutto ciò che usa "{{name}}" smetterà subito di funzionare.',
        revoked: 'Chiave API revocata',
        loadFailed: 'Impossibile caricare le chiavi API',
        createFailed: 'Impossibile creare la chiave API',
        revokeFailed: 'Impossibile revocare la chiave API',
        cancelBtn: 'Annulla'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistente',
      empty: "Apri il pannello e avvia un'attività da iniziare.",
      inputPlaceholder: 'Descrivi cosa vuoi che PicoClaw faccia',
      newConversation: 'Nuova conversazione',
      processing: 'In elaborazione...',
      agent: {
        defaultTitle: 'Assistente generale',
        defaultDescription: "Chat generale, ricerca e aiuto nell'area di lavoro.",
        kvmTitle: 'Controllo remoto',
        kvmDescription: "Gestisci l'host remoto tramite IronKVM.",
        switched: "Ruolo dell'agente cambiato",
        switchFailed: "Impossibile cambiare il ruolo dell'agente"
      },
      send: 'Invia',
      cancel: 'Annulla',
      status: {
        connecting: 'Connessione al gateway...',
        connected: 'Sessione PicoClaw connessa',
        disconnected: 'Sessione PicoClaw chiusa',
        stopped: 'Richiesta di interruzione inviata',
        runtimeStarted: 'Runtime PicoClaw avviato',
        runtimeStartFailed: 'Impossibile avviare il runtime PicoClaw',
        runtimeStopped: 'Runtime PicoClaw interrotto',
        runtimeStopFailed: 'Impossibile arrestare il runtime di PicoClaw',
        controlSwitchedToMCP: 'Controllo trasferito al servizio MCP esterno'
      },
      connection: {
        runtime: {
          checking: 'Controllo',
          restoring: 'Ripristino di PicoClaw',
          ready: 'Runtime pronto',
          stopped: 'Runtime interrotto',
          blockedByMCP: 'Il controllo MCP esterno è attivo',
          readyBlockedByMCP:
            "Il runtime è in esecuzione, ma un MCP esterno controlla ora l'input del dispositivo.",
          readyWithoutControl:
            'Il runtime è in esecuzione. Concedi a PicoClaw il controllo del dispositivo prima di riconnetterti.',
          unavailable: 'Runtime non disponibile',
          configError: 'Errore di configurazione'
        },
        transport: {
          connecting: 'Connessione',
          connected: 'Connesso',
          disconnected: 'Disconnesso',
          reconnect: 'Riconnetti',
          reconnectDescription: 'Riconnettiti alla sessione PicoClaw in esecuzione.',
          reconnectBlocked:
            'PicoClaw ha bisogno del controllo del dispositivo prima di riconnettersi.'
        },
        run: {
          idle: 'Inattivo',
          busy: 'Occupato'
        }
      },
      message: {
        toolAction: 'Azione',
        observation: 'Osservazione',
        screenshot: 'Schermata'
      },
      overlay: {
        locked: "PicoClaw sta controllando il dispositivo. L'immissione manuale è in pausa."
      },
      control: {
        picoclaw: 'Controllo dispositivo: PicoClaw',
        picoclawDescription:
          "PicoClaw può inviare input di tastiera e mouse. L'input manuale può andare in pausa.",
        mcp: 'Controllo dispositivo: MCP esterno',
        mcpDescription:
          "Il MCP esterno può scrivere sul dispositivo. PicoClaw non prenderà il controllo dell'input.",
        off: 'Controllo dispositivo: disattivato',
        offDescription:
          "L'IA non invierà input di tastiera o mouse. Il controllo manuale resta disponibile.",
        transitioning: 'Controllo dispositivo: cambio in corso',
        transitioningDescription: 'Il controllo del dispositivo si sta sincronizzando. Attendere.',
        grant: 'Concedi controllo',
        release: 'Rilascia',
        releasing: 'Rilascio...',
        switching: 'Cambio...',
        releasingLabel: 'Controllo dispositivo: rilascio in corso',
        releasingDescription:
          'Il controllo del dispositivo viene restituito. PicoClaw ha interrotto le scritture in corso.',
        granted: 'Controllo PicoClaw concesso',
        released: 'Controllo PicoClaw rilasciato',
        grantFailed: 'Impossibile concedere il controllo PicoClaw',
        releaseFailed: 'Impossibile rilasciare il controllo PicoClaw',
        grantConfirmTitle: 'Passare il controllo del dispositivo a PicoClaw?',
        grantConfirmDesc: 'Le scritture del dispositivo MCP esterno saranno interrotte.'
      },
      install: {
        install: 'Installa PicoClaw',
        installing: 'Installazione PicoClaw',
        success: 'PicoClaw installato correttamente',
        failed: 'Impossibile installare PicoClaw',
        uninstalling: 'Disinstallazione del runtime in corso...',
        uninstalled: 'Runtime disinstallato correttamente.',
        uninstallFailed: 'Disinstallazione non riuscita.',
        requiredTitle: 'PicoClaw non è installato',
        requiredDescription: 'Installa PicoClaw prima di avviare il runtime PicoClaw.',
        progressDescription: 'PicoClaw è in fase di download e installazione.',
        stages: {
          preparing: 'Preparazione',
          downloading: 'Download in corso',
          extracting: 'Estrazione',
          verifying: 'Verifica in corso',
          installing: 'Installazione in corso',
          installed: 'Installato',
          install_timeout: 'Timeout',
          install_failed: 'Non riuscito'
        }
      },
      model: {
        requiredTitle: 'È richiesta la configurazione del modello',
        requiredDescription:
          'Configura il modello PicoClaw prima di utilizzare la chat di PicoClaw.',
        docsTitle: 'Guida alla configurazione',
        docsDesc: 'Modelli e protocolli supportati',
        menuLabel: 'Configura modello',
        modelIdentifier: 'Identificatore del modello',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Chiave API',
        apiKeyPlaceholder: 'Inserisci la chiave API del modello',
        save: 'Salva',
        saving: 'Salvataggio',
        saved: 'Configurazione del modello salvata',
        saveFailed: 'Impossibile salvare la configurazione del modello',
        invalid: 'Identificatore del modello, API Base URL e chiave API sono obbligatori'
      },
      uninstall: {
        menuLabel: 'Disinstalla',
        confirmTitle: 'Disinstalla PicoClaw',
        confirmContent:
          "Sei sicuro di voler disinstallare PicoClaw? Ciò eliminerà l'eseguibile e tutti i file di configurazione.",
        confirmOk: 'Disinstalla',
        confirmCancel: 'Annulla'
      },
      history: {
        title: 'Cronologia',
        loading: 'Caricamento sessioni...',
        emptyTitle: 'Nessuna cronologia ancora',
        emptyDescription: 'Le sessioni PicoClaw precedenti verranno visualizzate qui.',
        loadFailed: 'Impossibile caricare la cronologia della sessione',
        deleteFailed: 'Impossibile eliminare la sessione',
        deleteConfirmTitle: 'Elimina sessione',
        deleteConfirmContent: 'Sei sicuro di voler eliminare "{{title}}"?',
        deleteConfirmOk: 'Elimina',
        deleteConfirmCancel: 'Annulla',
        messageCount_one: '{{count}} messaggio',
        messageCount_other: '{{count}} messaggi',
        messageCount: '{{count}} messaggi'
      },
      config: {
        startRuntime: 'Avvia PicoClaw',
        stopRuntime: 'Arresta PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Trasferire il controllo a PicoClaw?',
        enableConfirmDesc: 'L’avvio di PicoClaw disabiliterà il servizio MCP esterno.',
        enableConfirmOk: 'Avvia PicoClaw',
        enableConfirmCancel: 'Annulla',
        title: 'Avvia PicoClaw',
        description: "Avvia il runtime per iniziare a utilizzare l'assistente PicoClaw.",
        switchFromMCP: 'Passa a PicoClaw e avvia',
        takeoverAndStart: 'Prendi il controllo e avvia'
      }
    },
    error: {
      title: 'Si è verificato un problema',
      refresh: 'Aggiorna',
      panel: 'Questa parte della pagina ha smesso di funzionare',
      retry: 'Riprova'
    },
    fullscreen: {
      toggle: 'Attiva/disattiva schermo intero'
    },
    input: {
      disconnected: 'Tastiera e mouse non sono connessi',
      disconnectedTls:
        'Il browser ha rifiutato la connessione sicura che trasporta tastiera e mouse, cosa che fa senza chiedere. Il certificato generato da questo dispositivo non è ancora attendibile. Apri questo indirizzo in una nuova scheda, accetta il certificato, poi ricarica. Installare il certificato è la soluzione affidabile.',
      disconnectedNever:
        'Impossibile aprire la connessione che trasporta tastiera e mouse. Il resto della pagina funziona perché non la usa. Verifica che nulla tra te e il dispositivo la stia bloccando.',
      disconnectedDropped:
        'La connessione che trasporta tastiera e mouse si è interrotta e non è stata ripristinata. Si riconnette da sola dopo un riavvio; se il problema persiste, ricarica la pagina.',
      hidDisabled: "L'HID è disattivato su questo dispositivo (/boot/disable_hid).",
      keyFailed: 'Impossibile inviare il tasto.'
    },
    speaker: { title: 'Altoparlante', unmute: 'Riattiva audio', mute: 'Disattiva audio' },
    upstream: {
      check: 'Cerca aggiornamenti',
      updateTo: 'Aggiorna a {{version}}',
      confirm: 'Aggiornare {{name}} a {{version}}?',
      confirmDesc:
        'La nuova versione viene scaricata da GitHub e verificata con i checksum che pubblica. Se qualcosa non va, resta la versione attuale.',
      ok: 'Aggiorna',
      upToDate: 'Aggiornato',
      builtIn: 'integrata',
      checkFailed: 'Impossibile cercare aggiornamenti: {{error}}',
      unverifiable: 'La versione {{version}} non viene offerta: {{reason}}',
      inUse: 'Impossibile aggiornare ora: {{reason}}',
      running: 'Aggiornamento a {{version}}...',
      done: '{{name}} aggiornato a {{version}}',
      failed: "L'ultimo aggiornamento non è riuscito: {{error}}"
    },
    menu: {
      mediaNetboot: 'Impostazioni avvio di rete',
      collapse: 'Comprimi menu',
      expand: 'Espandi il menu',
      more: 'Altro',
      media: 'Supporti',
      tools: 'Strumenti',
      text: 'Testo',
      advanced: 'Avanzate',
      mediaMounted: 'Montato',
      mediaLibrary: 'Libreria',
      mediaBoot: 'Avvio',
      textToHost: "Verso l'host",
      textFromHost: "Dall'host"
    },
    ion: {
      checking: 'Controllo della memoria video prima di avviare lo streaming...',
      warn: 'La memoria video è scarsa. Un solo riavvio del server la esaurirebbe. Riavvia quando ti è comodo.',
      criticalTitle: 'Memoria video insufficiente per avviare lo streaming',
      criticalBody:
        "Avviare il video esaurirebbe la memoria riservata e arresterebbe il server. Tutte le altre funzioni restano disponibili, compresi il controllo dell'alimentazione e il riavvio. Solo un riavvio di IronKVM recupera questa memoria.",
      criticalContinue: 'Avvia comunque il video',
      criticalReboot: 'Riavvia IronKVM',
      criticalRebooting: 'Riavvio in corso...'
    }
  }
};

export default it;
