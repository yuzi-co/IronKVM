const it = {
  translation: {
    head: {
      desktop: 'Desktop Remoto',
      login: 'Accesso',
      changePassword: 'Cambia Password',
      terminal: 'Terminale',
      wifi: 'Wi-Fi'
    },
    auth: {
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
          'To reset the passwords, pressing and holding the BOOT button on the NanoKVM for 10 seconds.',
        reset2: 'Per i passaggi dettagliati, consulta questo documento:',
        reset3: 'Account web predefinito:',
        reset4: 'Account SSH predefinito:',
        change1: 'Tieni presente che questa azione modificherà le seguenti password:',
        change2: 'Password di accesso web',
        change3: 'Password root di sistema (password di accesso SSH)',
        change4: 'Per reimpostare le password, tieni premuto il pulsante BOOT sul NanoKVM.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Configura il Wi-Fi per NanoKVM',
      success: 'Please check the network status of NanoKVM and visit the new IP address.',
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
      }
    },
    screen: {
      scale: 'Scala',
      title: 'Schermo',
      video: 'Modalità video',
      videoDirectTips:
        'Abilita HTTPS in "Impostazioni > Dispositivo" per utilizzare questa modalità',
      resolution: 'Risoluzione',
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
      qualityLossless: 'Senza perdita',
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
      }
    },
    keyboard: {
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
      absoluteStalled: "L'host sta ignorando il mouse assoluto",
      absoluteStalledDesc:
        "L'host ha smesso di raccogliere i report del mouse assoluto, quindi i movimenti del puntatore vanno persi. La tastiera non è interessata. Ripristinare l'USB di solito risolve; la modalità relativa usa un endpoint diverso.",
      useRelative: 'Passa alla modalità relativa',
      direction: 'Direzione della rotellina',
      scrollUp: "Scorri verso l'alto",
      scrollDown: 'Scorri verso il basso',
      speed: 'Velocità della rotellina',
      fast: 'Veloce',
      slow: 'Lento',
      requestPointer:
        'Usando la modalità relativa. Clicca sul desktop per ottenere il puntatore del mouse.',
      resetHid: 'Reimposta HID',
      hidOnly: {
        title: 'Modalità solo HID',
        desc: 'Se il mouse e la tastiera smettono di rispondere e il ripristino di HID non aiuta, potrebbe trattarsi di un problema di compatibilità tra NanoKVM e il dispositivo. Prova ad abilitare la modalità HID-Only per una migliore compatibilità.',
        tip1: "L'abilitazione della modalità HID-Solo smonterà il disco U virtuale e la rete virtuale",
        tip2: "Nella modalità HID-Only, il montaggio dell'immagine è disabilitato",
        rebuild: 'Cambiare modalità ricostruisce la connessione USB. NanoKVM non si riavvia',
        enable: 'Abilita la modalità HID-Solo',
        disable: 'Disabilita la modalità HID-Solo'
      }
    },
    image: {
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
      tips: {
        title: 'Come caricare',
        usb1: 'Collega il NanoKVM al tuo computer tramite USB.',
        usb2: 'Assicurati che la Virtual Disk sia montata (Impostazioni - Virtual Disk).',
        usb3: 'Apri il disk virtuale sul tuo computer e copia il file immagine nella directory principale del disk.',
        scp1: 'Assicurati che il NanoKVM e il tuo computer siano sulla stessa rete locale.',
        scp2: 'Apri un terminale sul tuo computer e usa il comando SCP per caricare il file immagine nella directory /data del NanoKVM.',
        scp3: 'Esempio: scp il-tuo-percorso-immagine root@il-tuo-ip-nanokvm:/data',
        tfCard: 'Scheda TF',
        tf1: 'Questo metodo è supportato su sistemi Linux',
        tf2: 'Recupera la scheda TF dal NanoKVM (per la versione FULL, smonta prima il case).',
        tf3: 'Inserisci la scheda TF in un lettore di schede e collegala al tuo computer.',
        tf4: 'Copia il file immagine nella directory /data sulla scheda TF.',
        tf5: 'Inserisci la scheda TF nel NanoKVM.'
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
      close: 'Chiudi'
    },
    terminal: {
      title: 'Terminale',
      nanokvm: 'Terminale NanoKVM',
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
      title: 'Wake-on-LAN',
      sending: 'Invio comando...',
      sent: 'Comando inviato',
      input: 'Inserisci il MAC',
      ok: 'Ok'
    },
    download: {
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
      cancelFailed: 'Impossibile annullare il download'
    },
    power: {
      title: 'Accensione',
      showConfirm: 'Conferma',
      showConfirmTip: 'Le operazioni di alimentazione richiedono una conferma aggiuntiva',
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
      ledConnectedFailed: "Impossibile salvare l'impostazione del LED di accensione"
    },
    settings: {
      title: 'Impostazioni',
      mcp: {
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
        cancelBtn: 'Annulla'
      },
      redfish: {
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
        actionReset: 'Reset',
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
        failed: 'Operazione del watchdog non riuscita'
      },
      about: {
        title: 'Informazioni su NanoKVM',
        information: 'Informazioni',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versione Applicazione',
        applicationTip: 'Versione dell’applicazione web NanoKVM',
        image: 'Versione Immagine',
        imageTip: 'Versione dell’immagine di sistema NanoKVM',
        kernel: 'Versione del kernel',
        kernelTip: 'Versione del kernel Linux attualmente in esecuzione',
        deviceKey: 'Chiave Dispositivo',
        videoMemory: 'Memoria video',
        videoMemoryTip:
          'Memoria riservata alla cattura video. Non è condivisa con il resto del sistema.',
        videoMemoryGenerations_one:
          '{{count}} sessione precedente di NanoKVM sta trattenendo memoria video',
        videoMemoryGenerations_other:
          '{{count}} sessioni precedenti di NanoKVM stanno trattenendo memoria video',
        videoMemoryReboot: 'Riavvia per recuperarla.',
        community: 'Comunità',
        hostname: 'Nome host',
        hostnameUpdated: 'Nome host aggiornato. Riavviare per applicare.',
        ipType: {
          Wired: 'Cablato',
          Wireless: 'Senza fili',
          Other: 'Altro'
        }
      },
      appearance: {
        title: 'Aspetto',
        display: 'Schermo',
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
        ssh: {
          description: 'Abilita SSH accesso remoto',
          tip: "Imposta una password complessa prima dell'abilitazione (Account - Modifica password)"
        },
        advanced: 'Impostazioni avanzate',
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
        autostart: {
          title: 'Impostazioni script di avvio automatico',
          description:
            "Gestisce gli script che vengono eseguiti automaticamente all'avvio del sistema",
          new: 'Nuovo',
          deleteConfirm: 'Sei sicuro di voler eliminare questo file?',
          yes: 'Sì',
          no: 'No',
          scriptName: 'Nome script di avvio automatico',
          scriptContent: 'Contenuto script di avvio automatico',
          settings: 'Impostazioni'
        },
        hidOnly: 'HID-Solo modalità',
        hidOnlyDesc:
          'Smette di emulare i dispositivi virtuali, mantenendo solo il controllo di base HID',
        disk: 'Disco virtuale',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Rete virtuale',
        networkDesc: 'Monta la scheda di rete virtuale sull’host remoto',
        usbNetwork: {
          description:
            "Un collegamento di rete privato con l'host remoto tramite il cavo USB. L'host riceve un indirizzo senza gateway né DNS, quindi non può raggiungere la tua LAN tramite NanoKVM.",
          off: 'Disattivato',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (per host senza NCM)',
          rndis: 'RNDIS (non più offerto)',
          rndisNote: 'Questo collegamento usa RNDIS, che non è più offerto. Scegli NCM o ECM.',
          subnet: 'Sottorete',
          subnetDesc:
            "Una rete IPv4 privata, da /24 a /30. NanoKVM prende il primo indirizzo, l'host il secondo.",
          addresses: 'NanoKVM: {{board}}, host: {{host}}',
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
          "Presenta una porta seriale USB all'host remoto, per accedere a questo NanoKVM quando la rete non è raggiungibile",
        consoleTip:
          "Chiunque controlli l'host remoto ottiene un prompt di accesso a questo NanoKVM. Imposta una password complessa prima dell'abilitazione (Account - Modifica password).",
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
        rebootDesc: 'Sei sicuro di voler riavviare NanoKVM?',
        okBtn: 'Sì',
        cancelBtn: 'No'
      },
      network: {
        title: 'Rete',
        wifi: {
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
          waitingHttp: 'Ritorno a http. Ricarica questa pagina se non si apre da sola.'
        },
        ethernet: {
          title: 'Indirizzo IP',
          description: 'Configura come NanoKVM ottiene il suo indirizzo sulla rete cablata',
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
          applyTitle: "Cambiare l'indirizzo di NanoKVM?",
          applyWarning:
            "La connessione a questa pagina andrà persa. NanoKVM applica il nuovo indirizzo e attende {{seconds}} secondi che tu lo raggiunga a quell'indirizzo. Raggiungerlo mantiene la modifica. Se non lo raggiunge nulla, NanoKVM ripristina le impostazioni precedenti.",
          applyConfirm: 'Applica',
          applyCancel: 'Annulla',
          applyFailed: "Impossibile applicare l'indirizzo",
          trialTitle: 'In attesa di conferma',
          trialDhcp: 'NanoKVM sta chiedendo un indirizzo al DHCP.',
          trialStatic: 'NanoKVM ora si trova a {{address}}.',
          trialInstruction:
            'Apri NanoKVM al suo nuovo indirizzo e accedi se lo chiede. Raggiungerlo lì mantiene la modifica. Se nulla raggiunge NanoKVM entro {{seconds}} secondi, ripristina le impostazioni precedenti.',
          trialOpen: 'Apri il nuovo indirizzo',
          trialKeep: 'Mantieni queste impostazioni',
          trialKept: 'Il nuovo indirizzo è salvato',
          trialKeepFailed: 'Impossibile mantenere le impostazioni',
          trialGone: 'La modifica è già stata ripristinata. Riprova.',
          unsaved: 'Modifiche non salvate'
        },
        dns: {
          title: 'DNS',
          description: 'Configura i server DNS per NanoKVM',
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
        enable: 'Abilita {{name}}',
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
          tip: 'Se il demone resta a corto di memoria, prova ad abilitare la memoria di swap. Imposta la dimensione del file di swap a 256MB per impostazione predefinita, modificabile in "Impostazioni > Dispositivo".'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Riprova aggiornando la pagina o installa manualmente',
        download: 'Scarica il',
        package: 'pacchetto di installazione',
        unzip: 'e decomprimilo',
        upTailscale: 'Carica tailscale nella directory /usr/bin/ del NanoKVM',
        upTailscaled: 'Carica tailscaled nella directory /usr/sbin/ del NanoKVM',
        refresh: 'Aggiorna la pagina corrente',
        notLogin:
          'Il dispositivo non è ancora stato associato. Effettua il login e associa questo dispositivo al tuo account.',
        urlPeriod: 'Questo URL è valido per 10 minuti',
        login: 'Accedi',
        loginSuccess: 'Accesso riuscito',
        logout: 'Disconnetti',
        logoutDesc: 'Sei sicuro di voler uscire?'
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
        loginSuccess: 'Accesso riuscito',
        logout: 'Rimuovi registrazione',
        logoutDesc:
          'Rimuovere la registrazione elimina questo peer dal tuo account NetBird e ne cancella qui la configurazione. Per unirti di nuovo serve una chiave di configurazione o un accesso SSO, e il peer potrebbe ricevere un nuovo IP. Continuare?'
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
            'SHA-512 verifica soltanto che il pacchetto corrisponda al manifesto fornito da questo server. Non garantisce che il pacchetto sia una versione ufficiale di NanoKVM. Un server difettoso o dannoso può rendere inutilizzabile il dispositivo, causare la perdita di dati o compromettere il sistema.',
          confirm: 'Utilizza comunque',
          useSipeed: 'Usa il server ufficiale di Sipeed',
          previewDisabled:
            'Gli aggiornamenti in anteprima non sono disponibili quando è attivo un server di aggiornamento personalizzato.'
        },
        offline: {
          title: 'Aggiornamenti offline',
          desc: 'Aggiornamento tramite pacchetto di installazione locale',
          upload: 'Carica',
          checksumPlaceholder: 'Checksum SHA-256 (facoltativo)',
          invalidChecksum: 'Il checksum SHA-256 deve contenere 64 caratteri esadecimali.',
          checksumMismatch:
            'La verifica SHA-256 non è riuscita. Il pacchetto potrebbe essere danneggiato.',
          invalidName: 'Formato nome file non valido. Si prega di scaricare dalle versioni GitHub.',
          updateFailed: 'Aggiornamento fallito. Riprova.'
        }
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
        kvmDescription: "Gestisci l'host remoto tramite NanoKVM.",
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
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime pronto',
          stopped: 'Runtime interrotto',
          blockedByMCP: 'Il controllo MCP esterno è attivo',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime non disponibile',
          configError: 'Errore di configurazione'
        },
        transport: {
          connecting: 'Connessione',
          connected: 'Connesso',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
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
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Controllo dispositivo: MCP esterno',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Controllo dispositivo: disattivato',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Concedi controllo',
        release: 'Rilascia',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
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
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
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
    menu: {
      collapse: 'Comprimi menu',
      expand: 'Espandi il menu'
    },
    ion: {
      checking: 'Controllo della memoria video prima di avviare lo streaming...',
      warn: 'La memoria video è scarsa. Un solo riavvio del server la esaurirebbe. Riavvia quando ti è comodo.',
      criticalTitle: 'Memoria video insufficiente per avviare lo streaming',
      criticalBody:
        "Avviare il video esaurirebbe la memoria riservata e arresterebbe il server. Tutte le altre funzioni restano disponibili, compresi il controllo dell'alimentazione e il riavvio. Solo un riavvio di NanoKVM recupera questa memoria.",
      criticalContinue: 'Avvia comunque il video',
      criticalReboot: 'Riavvia NanoKVM',
      criticalRebooting: 'Riavvio in corso...'
    }
  }
};

export default it;
