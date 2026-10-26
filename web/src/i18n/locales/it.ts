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
      share: {
        opening: 'Apertura dello schermo condiviso...',
        failed: 'Questo link non è valido, è scaduto o è stato revocato.',
        toLogin: 'Accedi invece'
      },
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
        verifyBtn: 'Verifica',
        tooManyAttempts: 'Troppe password errate. Attendi un momento e riprova.'
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
      streamInfo: {
        toggle: 'Info stream',
        toggleTip:
          'Mostra sotto l\'immagine una riga con connessione, risoluzione, frequenza dei fotogrammi, bitrate e latenza dello stream, come li misura questo browser.',
        title: 'Stato dello stream',
        connecting: 'Connessione',
        connected: 'Connesso',
        reconnecting: 'Riconnessione',
        failed: 'Non riuscito',
        paused: 'In pausa',
        noSignal: 'Nessun segnale',
        source: 'Sorgente',
        sourceTip: 'La modalità dell\'ingresso HDMI catturata dal dispositivo',
        stream: 'Stream',
        streamTip: 'La dimensione dell\'immagine come arriva in questo browser',
        received: 'ricevuti',
        decoded: 'decodificati',
        dropped: 'scartati',
        board: 'dispositivo',
        boardTip: 'Fotogrammi al secondo catturati e codificati dal dispositivo',
        bitrateTip: 'Dati video ricevuti da questo browser',
        latency: 'Latenza',
        latencyCapture:
          'Dalla cattura sul dispositivo alla visualizzazione qui, in base all\'ora di cattura che il server invia con ogni fotogramma.',
        latencyEncoder:
          'Dall\'uscita dall\'encoder al disegno qui, misurata sull\'orologio del dispositivo. Cattura e codifica non sono incluse.',
        latencyEstimate:
          'Stima: metà del tempo di andata e ritorno di rete più buffer e decodifica in questo browser. Cattura e codifica sul dispositivo non sono incluse, quindi la latenza reale è maggiore.',
        latencyEstimateSum:
          'Somma di parti misurate: {{board}} ms sul dispositivo dalla cattura all\'invio, {{network}} ms di percorso fin qui (metà del tempo di andata e ritorno di rete), {{receive}} ms di buffer e decodifica in questo browser e {{display}} ms di attesa del display (mezzo aggiornamento).',
        latencyClockOffset:
          'Corretto per la differenza tra l\'orologio del dispositivo e quello di questo computer, misurata in {{offset}} ms (dispositivo meno questo computer).',
        latencyClockAssumed:
          'Si presume che l\'orologio del dispositivo coincida con quello di questo computer.',
        latencyArrival:
          'I fotogrammi arrivano qui {{ms}} ms dopo l\'encoder; il resto è decodifica e disegno in questo browser.',
        rttTip: 'Andata e ritorno di rete verso il dispositivo',
        loss: 'Perdita',
        lossTip: 'Pacchetti persi nell\'ultimo secondo',
        transportTip: 'Come il video raggiunge questo browser'
      },
      stream: {
        ok: 'immagine OK',
        noSignal: 'nessun segnale',
        failed: 'stream non riuscito'
      },
      codecNoWebrtcHevc: 'Questo browser non può ricevere H.265 tramite WebRTC',
      codecNoHevc: 'Questo browser non può decodificare H.265',
      codecNote:
        'La scheda ha un solo encoder, quindi questo cambia lo stream per tutti gli spettatori.',
      codec: 'Codec',
      updateFailed: "L'impostazione non è stata applicata",
      scale: 'Scala',
      pan: "Trascina per spostare l'immagine ingrandita, tocca due volte per centrarla",
      title: 'Schermo',
      video: 'Modalità video',
      videoDirectTips:
        'Abilita HTTPS in "Impostazioni > Dispositivo" per utilizzare questa modalità',
      resolution: 'Risoluzione',
      aspect: 'Proporzioni',
      aspectKeep: 'Automatico (mantieni le proporzioni della sorgente)',
      aspectStretch: 'Estendi alla risoluzione',
      aspectTips:
        "Automatico mantiene la forma dello schermo dell'host all'altezza scelta: uno schermo 1920x1200 viene inviato come 1728x1080. Estendi riempie la risoluzione scelta e deforma gli schermi che non sono 16:9.",
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
      onScreen: 'Tastiera su schermo',
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
      autoTap: {
        title: 'Pressione automatica',
        tip: 'Preme un tasto più e più volte per aprire un menu del firmware o di avvio. Il dispositivo continua a premere anche quando questa scheda viene chiusa.',
        key: 'Tasto',
        otherKey: 'Altro tasto',
        interval: 'Intervallo (ms)',
        timeout: 'Ferma dopo (s)',
        firstPress: 'Prima pressione',
        firstPressNone: 'Nulla',
        firstPressPower: 'Pulsante di accensione',
        firstPressReset: 'Pulsante di reset',
        stopOnInput: 'Ferma al mio prossimo tasto o clic',
        stopOnVideo: 'Ferma quando cambia la modalità video',
        stopOnVideoTip:
          'Un menu del firmware ha spesso una risoluzione diversa da quella del sistema. Un riavvio può cambiare la modalità più di una volta, e alcuni firmware mantengono la stessa modalità.',
        start: 'Avvia',
        stop: 'Stop'
      },
      macro: {
        title: 'Macro',
        empty: 'Nessuna macro salvata',
        record: 'Registra una macro',
        recordTip:
          'Registra i tuoi tasti, il mouse, le pause tra di essi e le pressioni di accensione o reset, finché non fai clic su Fine.',
        play: 'Riproduci',
        edit: 'Modifica',
        delete: 'Elimina',
        deleteConfirm: 'Eliminare la macro "{{name}}"?',
        recording: 'Registrazione di una macro. Passaggi: {{count}}',
        recordingFull: 'La macro ha raggiunto il limite di passaggi.',
        finish: 'Fine',
        discard: 'Scarta',
        editorTitle: 'Macro',
        name: 'Nome',
        namePlaceholder: 'Nome della macro',
        summary: 'Passaggi: {{count}}. Durata: circa {{seconds}} s',
        addPause: 'Aggiungi una pausa',
        addPower: 'Aggiungi una pressione di accensione',
        addReset: 'Aggiungi una pressione di reset',
        removeStep: 'Rimuovi il passaggio',
        playNow: 'Riproduci ora',
        save: 'Salva',
        saved: 'Macro salvata',
        adminOnly: 'Solo un amministratore può salvare le macro.',
        noSteps: 'La macro non ha passaggi.',
        noName: 'Dai un nome alla macro.',
        stepKeys: 'Premi {{keys}}',
        stepRelease: 'Rilascia i tasti',
        stepPointerAt: 'Puntatore a {{x}}%, {{y}}%',
        stepPointerBy: 'Spostamento del puntatore {{x}}, {{y}}',
        stepButtons: 'pulsanti: {{buttons}}',
        stepNoButtons: 'nessun pulsante',
        stepScroll: 'scorrimento {{amount}}',
        stepPause: 'Pausa (ms)',
        stepPower: 'Pulsante di accensione, {{ms}} ms',
        stepReset: 'Pulsante di reset, {{ms}} ms',
        buttonLeft: 'sinistro',
        buttonRight: 'destro',
        buttonMiddle: 'centrale'
      },
      sequence: {
        autoTapRunning: 'Pressione automatica {{name}}. Pressioni: {{taps}}',
        macroRunning: 'Macro {{name}}. Passaggio {{step}} di {{total}}',
        elapsed: '{{elapsed}} s di {{timeout}} s',
        stop: 'Stop',
        alreadyRunning: 'Una pressione automatica o una macro è già in esecuzione. Fermala prima.',
        endedAutoTap: 'Pressione automatica {{name}}: {{reason}}',
        endedMacro: 'Macro {{name}}: {{reason}}',
        reasons: {
          finished: 'completata',
          stopped: 'fermata',
          timeout: 'fermata al limite di tempo',
          input: 'fermata dal tuo tasto o clic',
          video: 'fermata, la modalità video è cambiata',
          shutdown: 'fermata, il server si è riavviato',
          busy: 'non riuscita, un altro controller sta usando la tastiera',
          hid_error: "non riuscita, l'host non ha accettato l'input",
          atx_error: 'non riuscita, la pressione di accensione o reset non ha funzionato'
        }
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
          ko: 'Coreano',
          am: 'Armeno',
          ar: 'Arabo',
          bepo: 'Francese (BÉPO)',
          cz: 'Ceco',
          da: 'Danese',
          deCh: 'Tedesco (Svizzera)',
          enUsAltgrIntl: 'Inglese (Stati Uniti, AltGr-Intl)',
          enUsColemak: 'Inglese (Colemak)',
          et: 'Estone',
          fo: 'Faroese',
          frBe: 'Francese (Belgio)',
          frCa: 'Francese (Canada)',
          frCh: 'Francese (Svizzera)',
          hr: 'Croato',
          hu: 'Ungherese',
          is: 'Islandese',
          lt: 'Lituano',
          lv: 'Lettone',
          mk: 'Macedone',
          nl: 'Olandese',
          no: 'Norvegese',
          pl: 'Polacco',
          pt: 'Portoghese (Portogallo)',
          sl: 'Sloveno',
          th: 'Thailandese',
          tr: 'Turco'
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
      keyJiggler: 'Pressione automatica di un tasto',
      keyJigglerF15: 'Tasto F15',
      keyJigglerShift: 'Tasto Maiusc',
      keyJigglerCtrl: 'Tasto Ctrl',
      keyJigglerF15Tip: 'F15 è il meno invasivo: nessun sistema o app comune lo usa',
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
      ok: 'Ok',
      add: 'Aggiungi dispositivo',
      edit: 'Modifica dispositivo',
      name: 'Nome',
      mac: 'Indirizzo MAC',
      scan: 'Cerca nella rete',
      scanFailed: 'Impossibile analizzare la rete',
      scanEmpty: 'Nessun host trovato',
      scanPick: 'Scegli un host',
      ip: "Indirizzo IP dell'host (per il controllo di accensione, facoltativo)",
      broadcast: 'Indirizzo di broadcast o sottorete (facoltativo)',
      port: 'Pacchetto di risveglio',
      portFrame: 'Frame Ethernet (predefinito)',
      secureOn: 'Password SecureOn (facoltativa)',
      secureOnKept: 'Salvata. Lascia vuoto per mantenerla',
      secureOnRemove: 'Rimuovi la password',
      secureOnKeep: 'Mantieni la password',
      invalidMac: 'Inserisci un indirizzo MAC valido',
      invalidIp: 'Inserisci un indirizzo IPv4 valido',
      invalidBroadcast: 'Inserisci un indirizzo IPv4 o una sottorete, ad esempio 192.168.1.0/24',
      invalidSecureOn: 'La password è di sei byte in esadecimale, ad esempio AA:BB:CC:DD:EE:FF',
      duplicate: 'Esiste già un dispositivo con questo indirizzo MAC o nome',
      saveFailed: 'Impossibile salvare il dispositivo',
      save: 'Salva',
      cancel: 'Annulla',
      up: 'Acceso',
      down: 'Spento',
      checking: 'Controllo...'
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
        boot: 'Avvio e supporti',
        browser: 'Questo browser',
        search: "Cerca un'impostazione",
        noMatch: 'Nessuna impostazione corrisponde',
        locked:
          "Un'operazione è in corso. Le altre pagine e la chiusura non sono disponibili finché non termina.",
        vpnProvider: 'Provider VPN'
      },
      audit: {
        title: 'Registro di audit',
        log: 'Eventi',
        description:
          'Chi ha effettuato l\'accesso, preso il controllo, premuto l\'accensione, cambiato supporti o impostazioni, e cosa hanno fatto i client IA. Quando il registro è pieno, gli eventi più vecchi vengono eliminati.',
        refresh: 'Aggiorna',
        export: 'Esporta',
        time: 'Ora',
        who: 'Chi',
        address: 'Indirizzo',
        category: 'Categoria',
        action: 'Azione',
        target: 'Destinazione',
        outcome: 'Esito',
        tier: 'Livello: {{tier}}',
        empty: 'Nessun evento',
        searchPlaceholder: 'Cerca azione, destinazione o dettaglio',
        loadFailed: 'Impossibile caricare il registro di audit.',
        exportFailed: 'Impossibile esportare il registro di audit.',
        saveFailed: 'Impossibile salvare l\'impostazione.',
        syslog: 'Invia gli eventi a syslog',
        syslogDescription:
          'Invia ogni evento anche al registro di sistema, che lo inoltra al collettore syslog remoto se ne è impostato uno in Rete.',
        storage: 'Conservato in {{directory}}, al massimo {{size}} KB.',
        categories: {
          auth: 'Accesso',
          control: 'Controllo',
          power: 'Alimentazione',
          media: 'Supporti',
          settings: 'Impostazioni',
          macro: 'Macro',
          system: 'Sistema',
          agent: 'Agenti IA',
        },
        outcomes: {
          ok: 'Riuscito',
          fail: 'Non riuscito',
          denied: 'Negato'
        }
      },
      mcp: {
        tiers: {
          title: 'Livelli di capacità',
          description:
            'Scegli cosa possono fare i client MCP e l\'agente integrato. Ogni chiamata viene registrata nel registro di audit con il suo livello.',
          mcp: 'Client MCP',
          agent: 'Agente integrato',
          observe: 'Osservare',
          observeDesc: 'Schermate e stato',
          input: 'Input',
          inputDesc: 'Tastiera e mouse',
          power: 'Alimentazione',
          powerDesc: 'Premere i pulsanti di accensione e reset ATX',
          media: 'Supporti',
          mediaDesc: 'Inserire ed espellere supporti virtuali',
          admin: 'Amministrazione',
          adminDesc:
            'Modificare le impostazioni del dispositivo (nessuno strumento usa ancora questo livello)',
          agentNote:
            'L\'agente integrato non ha ancora strumenti di alimentazione o supporti, quindi questi due livelli valgono solo per i client MCP.',
          loadFailed: 'Impossibile caricare i livelli di capacità.',
          saveFailed: 'Impossibile salvare i livelli di capacità.'
        },
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
          "sshd accetta ancora le password. Spegni e riaccendi il server SSH per applicare l'impostazione.",
        changePort: 'Cambia',
        portConfirm: 'Cambiare la porta SSH in {{port}}?',
        portConfirmDesc:
          'Le sessioni SSH attuali restano aperte. Le nuove connessioni devono usare la porta {{port}}. Assicurati che il firewall lo consenta.',
        portChanged: 'Porta SSH cambiata in {{port}}',
        portInvalid: 'Inserisci una porta da 1 a 65535.',
        portReserved: "IronKVM usa già questa porta. Scegline un'altra.",
        portInUse: "Un altro programma sull'IronKVM è già in ascolto su questa porta.",
        portNotHonoured:
          'Lo sshd di questa immagine non legge questa impostazione, quindi la porta resta invariata.'
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
      notifications: {
        title: 'Notifiche',
        service: 'Invia notifiche',
        serviceDesc:
          "Invia un avviso, con un'immagine dello schermo, quando succede qualcosa all'host o a questo KVM. Disattivato per impostazione predefinita.",
        snapshot: "Allega un'istantanea",
        snapshotDesc:
          "Un JPEG dello schermo accompagna l'avviso dove il canale può trasportarlo. Non viene salvato.",
        events: 'Eventi',
        eventsDesc:
          'Scegli cosa viene inviato e il tempo minimo tra due avvisi dello stesso tipo. Gli avvisi trattenuti da questo intervallo vengono conteggiati e riportati con il successivo.',
        event: {
          signalLost: 'Segnale HDMI perso',
          signalRestored: 'Segnale HDMI ripristinato',
          watchdogAction: 'Il watchdog ha premuto reset o accensione',
          loginFailures: 'Accessi falliti ripetuti',
          atxPress: 'Pulsante di accensione o reset premuto',
          failsafeEntered: 'Modalità failsafe: il server si è avviato senza video',
          hostLogMatch: "Frase nel log dell'host, come un kernel panic"
        },
        cooldown: 'Tempo minimo tra due avvisi',
        seconds: 'secondi',
        loginWithin: 'accessi falliti entro',
        channels: 'Canali',
        webhook: 'Webhook',
        webhookDesc: "Invia l'evento come JSON a un indirizzo.",
        webhookUrl: 'Indirizzo',
        webhookSnapshot: "Aggiungi l'istantanea al JSON come base64",
        ntfy: 'ntfy',
        ntfyDesc: 'Pubblica su un argomento di ntfy.sh o sul tuo server.',
        ntfyServer: 'Server',
        ntfyTopic: 'Argomento',
        ntfyToken: 'Token di accesso (facoltativo)',
        telegram: 'Telegram',
        telegramDesc: 'Invia tramite un bot creato con BotFather.',
        botToken: 'Token del bot',
        chatId: 'Chat ID',
        savedSecret: 'Salvato. Digita per sostituirlo.',
        savedHint: 'Salvato per {{hint}}. Digita per sostituirlo.',
        willRemove: 'Verrà rimosso al salvataggio',
        remove: 'Rimuovi',
        sendTest: 'Invia test',
        sendTestAll: 'Invia test a ogni canale',
        saveFirst: 'Salva prima le modifiche',
        testOk: 'Test inviato su {{channel}}',
        testChannelFailed: '{{channel}}: {{error}}',
        testFailed: 'Impossibile inviare il test',
        save: 'Salva',
        saved: 'Salvato',
        failed: 'Impostazioni delle notifiche non riuscite',
        recent: 'Avvisi recenti',
        refresh: 'Aggiorna',
        noRecent: 'Non è stato ancora inviato nulla.',
        heldBack: 'Trattenuti',
        lastOk: 'inviato',
        lastFailed: 'non riuscito'
      },
      hostLog: {
        title: "Log dell'host",
        intro:
          "Ricevi i log dell'host gestito tramite la rete: syslog e netconsole del kernel. Entrambi sono disattivati finché non li attivi. Le righe restano in memoria e, se vuoi, anche su disco.",
        syslog: 'Syslog (UDP)',
        syslogDesc: "RFC 3164 e RFC 5424. Punta rsyslog o syslog-ng dell'host a questo KVM.",
        netconsole: 'Netconsole (UDP)',
        netconsoleDesc:
          "Messaggi del kernel, anche mentre i dischi dell'host si guastano. Formati semplice ed esteso.",
        port: 'Porta',
        state: {
          off: 'Disattivato',
          listening: 'In ascolto',
          waiting: 'In attesa',
          error: 'Errore'
        },
        stateListening: 'Su {{addr}}',
        stateWaiting: 'In attesa del collegamento di rete USB',
        bind: 'Ascolta su',
        bindDesc: "L'opzione USB sente solo l'host collegato con il cavo a questo KVM.",
        bindAll: 'Tutte le interfacce',
        bindUsb: 'Solo il collegamento di rete USB',
        sources: 'Mittenti accettati',
        sourcesDesc:
          "Indirizzi o reti, uno per riga. Se vuoto, si sente solo l'host sul collegamento USB (o l'indirizzo di ping del watchdog) quando è noto, altrimenti qualsiasi mittente.",
        inForce: 'In vigore: {{list}}',
        inForceAny: 'In vigore: qualsiasi mittente',
        memory: 'Buffer in memoria',
        memoryDesc:
          'Vengono conservate le righe più recenti che ci stanno. Le più vecchie vengono scartate.',
        disk: 'Conserva il log anche su disco',
        diskDesc:
          'Sopravvive a un riavvio. Scrive sulla scheda SD, quindi è disattivato per impostazione predefinita.',
        diskLimit: 'Limite su disco',
        patterns: 'Frasi di avviso',
        patternsDesc:
          'Espressioni regolari, una per riga, maiuscole o minuscole. Una riga corrispondente invia una notifica. Se vuoto, si usano quelle predefinite: kernel panic, oops, BUG e out of memory.',
        save: 'Salva',
        saved: 'Salvato',
        failed: "Richiesta del log dell'host non riuscita",
        stats:
          'Ricevute {{received}} righe, ignorate {{filtered}} da altri mittenti, scartate {{dropped}} perché troppo veloci. In memoria {{count}} righe.',
        live: 'Log in tempo reale',
        paused: 'In pausa',
        connected: 'In diretta',
        reconnecting: 'Riconnessione',
        search: 'Cerca parole',
        severityAll: 'Tutti i livelli',
        severityError: 'Errori e superiori',
        severityWarning: 'Avvisi e superiori',
        severityInfo: 'Info e superiori',
        pause: 'Pausa',
        resume: 'Riprendi',
        download: 'Scarica',
        clear: 'Svuota',
        cancel: 'Annulla',
        empty: 'Ancora nessuna riga.',
        clearConfirmTitle: "Svuotare il log dell'host?",
        clearConfirmDesc: 'Le righe conservate in memoria e su disco vengono eliminate.'
      },
      nut: {
        title: 'Client NUT',
        service: 'Monitoraggio UPS',
        serviceDesc: 'Legge un UPS da un server NUT in rete. Con batteria scarica o spegnimento forzato preme il pulsante di accensione dell\'host per spegnerlo, e lo preme di nuovo al ritorno della rete elettrica.',
        notice: 'Il KVM deve restare alimentato durante un\'interruzione e l\'header ATX deve essere collegato. Il protocollo NUT non è cifrato: utente e password viaggiano in chiaro sulla rete.',
        needsLed: 'Il client NUT richiede che "LED di alimentazione collegato" sia attivo, per sapere se l\'host è acceso.',
        ledHint: '"LED di alimentazione collegato" è disattivato, quindi il client non preme alcun pulsante.',
        host: 'Server NUT',
        hostDesc: 'La macchina che esegue upsd: un NAS, un router, una scheda di rete dell\'UPS o l\'host stesso.',
        hostInvalid: 'Inserisci un nome host o un indirizzo IP.',
        port: 'Porta',
        upsName: 'Nome UPS',
        username: 'Utente',
        credentialsDesc: 'Facoltativo. Con utente e password il KVM accede come secondario, così il primario attende lo spegnimento dell\'host prima di tagliare l\'uscita dell\'UPS.',
        password: 'Password',
        passwordSet: 'Una password è memorizzata.',
        passwordKeep: 'Invariata',
        passwordClear: 'Rimuovi password',
        onBattery: 'Spegni dopo',
        onBatteryDesc: 'Minuti a batteria prima di chiedere lo spegnimento dell\'host. 0 lo disattiva; con batteria scarica e spegnimento forzato l\'host si spegne sempre.',
        forceOff: 'Forza lo spegnimento dopo',
        forceOffDesc: 'Minuti dopo la pressione di spegnimento prima di tenere premuto il pulsante di accensione per 5 secondi. 0 lo disattiva.',
        powerOn: 'Accendi al ritorno della rete elettrica',
        powerOnDesc: 'Preme il pulsante di accensione quando l\'UPS torna alimentato dalla rete e l\'host è spento dopo un\'interruzione. Un host spento di proposito resta com\'è.',
        minCharge: 'Carica minima',
        minChargeDesc: 'Carica della batteria che l\'UPS deve prima raggiungere. 0 significa qualsiasi. L\'UPS deve riportare battery.charge.',
        delay: 'Ritardo',
        delayDesc: 'Per quanto tempo la rete elettrica deve essere tornata prima della pressione.',
        minutes: 'min',
        secondsUnit: 's',
        save: 'Salva',
        saved: 'Salvato',
        failed: 'Impossibile caricare o salvare le impostazioni NUT',
        state: 'Stato',
        status: {
          off: 'Disattivato',
          unconfigured: 'Non configurato',
          connecting: 'Connessione',
          ok: 'Connesso',
          unreachable: 'Non raggiungibile'
        },
        upsStateLabel: 'UPS',
        upsState: {
          online: 'Alimentato da rete',
          onBattery: 'A batteria',
          lowBattery: 'Batteria scarica',
          shutdown: 'Spegnimento',
          unknown: 'Sconosciuto'
        },
        charge: 'Carica',
        runtime: 'Autonomia',
        onBatteryFor: 'A batteria da',
        hostPower: 'Alimentazione host',
        hostState: {
          on: 'Acceso',
          off: 'Spento',
          unknown: 'Sconosciuto'
        },
        login: 'Accesso',
        loginSecondary: 'Secondario: il primario attende',
        loginReadOnly: 'Sola lettura',
        loginRefused: 'Rifiutato ({{code}})',
        activity: 'Attività',
        phase: {
          idle: 'Inattivo',
          shuttingDown: 'Spegnimento dell\'host',
          waitingForMains: 'In attesa della rete elettrica per accendere',
          needsLed: 'Richiede il LED di alimentazione'
        },
        powerOnIn: 'Accensione tra',
        lastPoll: 'Ultima interrogazione',
        never: 'Mai',
        duration: '{{minutes}} min {{seconds}} s',
        events: 'Eventi',
        noEvents: 'Ancora nulla.',
        event: {
          unreachable: 'Server NUT non raggiungibile',
          reachable: 'Server NUT di nuovo raggiungibile',
          onBattery: 'UPS a batteria',
          mainsBack: 'La rete elettrica è tornata',
          shutdown: 'Spegnimento host premuto',
          forceOff: 'Host spento forzatamente',
          hostOff: 'L\'host è spento',
          powerOn: 'Accensione host premuta',
          pressFailed: 'Pressione del pulsante non riuscita'
        },
        reason: {
          fsd: 'spegnimento forzato',
          lowBattery: 'batteria scarica',
          onBattery: 'tempo a batteria'
        },
        barTip: 'UPS letto dal server NUT: stato, carica della batteria e autonomia',
        runtimeShort: '{{minutes}} min'
      },
      backup: {
        title: 'Backup',
        settingsTitle: 'File delle impostazioni',
        settingsDesc: 'Esporta le impostazioni di questo dispositivo in un file. Importa un file da questo dispositivo o da un altro.',
        excluded: 'Il file non contiene password, chiavi, token o certificati. Le impostazioni di rete, gli account e SSH non sono nel file e un’importazione non li modifica.',
        export: 'Esporta',
        import: 'Importa',
        exportFailed: 'Esportazione non riuscita',
        previewTitle: 'Importa impostazioni',
        previewSource: 'File da {{host}}, versione {{version}}',
        previewFailed: 'Impossibile leggere il file',
        notJson: 'Il file non è un JSON valido.',
        notSettings: 'Il file non è un file di impostazioni IronKVM.',
        newer: 'Il file proviene da un firmware più recente. Aggiorna prima questo dispositivo.',
        tooLarge: 'Il file è troppo grande.',
        noChanges: 'Il file non modifica nulla su questo dispositivo.',
        items: 'Elementi: {{n}}',
        importSelected: 'Importa',
        cancel: 'Annulla',
        importFailed: 'Importazione non riuscita. Nulla è stato modificato.',
        imported: 'Impostazioni importate',
        rebootTitle: 'Riavviare ora?',
        rebootDesc: 'Il dispositivo applica le impostazioni importate al prossimo avvio.',
        rebootNow: 'Riavvia ora',
        rebootLater: 'Più tardi',
        rebootFailed: 'Riavvio non riuscito',
        diagnosticsTitle: 'Pacchetto di diagnostica',
        diagnosticsDesc: 'Un archivio con le versioni, i log e lo stato di rete, USB e video, per una richiesta di assistenza. Password, chiavi e token vengono rimossi.',
        download: 'Scarica',
        downloadFailed: 'Impossibile creare il pacchetto',
        status: {
          change: 'Modifiche',
          same: 'Nessuna modifica',
          invalid: 'Non valido',
          unknown: 'Sconosciuto'
        },
        sections: {
          logging: 'Registrazione',
          loginSecurity: 'Blocco accesso',
          sessions: 'Sessioni',
          powerLed: 'LED di alimentazione',
          redfish: 'Redfish',
          ipmi: 'IPMI',
          watchdog: 'Watchdog',
          nut: 'Client NUT',
          netboot: 'Avvio di rete',
          vnc: 'VNC',
          webrtc: 'WebRTC',
          picoclaw: 'Screenshot IA',
          hostLog: 'Log dell’host',
          webTitle: 'Titolo web',
          leaderKey: 'Tasto leader',
          keyboardLayout: "Layout di tastiera dell'host",
          shortcuts: 'Scorciatoie da tastiera',
          macros: 'Macro',
          mouseJiggler: 'Jiggler del mouse',
          keyJiggler: 'Jiggler dei tasti',
          oled: 'Display OLED',
          hdmi: 'Timeout inattività HDMI',
          roomMic: 'Microfono ambiente',
          notifications: 'Regole di notifica'
        }
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
      media: {
        title: 'Supporti virtuali',
        description:
          'Configurazione della finestra Supporti nella barra degli strumenti. Montare le immagini, aggiungerle e scegliere il set di Ventoy si fa nella finestra.',
        ejectFirst:
          "Il disco Ventoy è in un'unità. Espellilo nella finestra Supporti per disinstallare."
      },
      netboot: {
        title: 'Avvio di rete',
        isoDownload: 'Scarica',
        updateFiles: 'File di avvio (avvio di rete)',
        updateIso: 'ISO del menu di avvio (CD virtuale)',
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
        basedOn: 'basato su NanoKVM {{version}}',
        licenses: 'Licenze',
        licensesDescription: 'Licenze del software di questa immagine',
        licensesUnavailable: 'Questa immagine non contiene un manifesto delle licenze.',
        licenseTexts: 'Testi delle licenze ({{count}})'
      },
      preferences: {
        title: 'Preferenze'
      },
      performance: {
        title: 'Prestazioni',
        memory: {
          title: 'Memoria',
          description: 'RAM, swap e cosa li usa. Si aggiorna ogni pochi secondi.',
          ram: 'RAM',
          of: '{{used}} di {{total}}',
          available: '{{available}} disponibili',
          availableLow:
            'Solo {{available}} disponibili. I servizi possono rallentare o essere fermati.',
          swap: 'Swap',
          swapFile: 'File di swap',
          zram: 'Swap compresso',
          zramRam: '{{ram}} in RAM',
          consumers: 'Principali consumatori',
          addons: 'Componenti aggiuntivi',
          addonsTip:
            'Il gruppo di memoria in cui girano Tailscale e NetBird, rispetto al suo limite.',
          video: 'Memoria video',
          videoTip:
            'Memoria riservata alla cattura video. Non è condivisa con il resto del sistema.',
          videoGenerations_one:
            '{{count}} sessione precedente di IronKVM sta trattenendo memoria video',
          videoGenerations_other:
            '{{count}} sessioni precedenti di IronKVM stanno trattenendo memoria video',
          videoReboot: 'Riavvia per recuperarla.',
          videoUnmeasured: '{{size}} riservati, uso non misurabile',
          largestFree: 'Blocco libero più grande {{size}}',
          peak: 'picco {{size}}',
          cma: 'Memoria contigua (CMA)',
          cmaTip:
            'Memoria che il kernel riserva ai buffer che devono essere contigui, come i fotogrammi video tra lo scaler e gli encoder. Finché non è assegnata, il resto del sistema può prenderla in prestito.'
        }
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
        hostLayout: "Layout di tastiera dell'host",
        hostLayoutDesc:
          "Il layout con cui digita l'host. Gli agenti IA digitano con esso, e l'incolla parte da esso in un browser che non ne ha scelto uno.",
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
          frontPanel: 'Pannello frontale',
          roomMic: 'Microfono della stanza'
        },
        roomMic: {
          allow: 'Consenti il microfono della stanza',
          description:
            'Chi guarda può ascoltare la stanza tramite il microfono integrato del KVM. Chiunque guardi vede quando è acceso.',
          unavailable: 'Non disponibile con questo kernel',
          gain: 'Guadagno del microfono'
        },
        cpuFreq: {
          title: 'Frequenza CPU',
          description: 'Imposta la frequenza della CPU applicata al prossimo avvio',
          tip: 'La CPU si avvia a 850 MHz ed è specificata per 1000 MHz. Un nuovo valore si applica al prossimo avvio, non mentre il sistema è in funzione. 1000 MHz rientra nelle specifiche; la temperatura resta ben entro i limiti con entrambe le impostazioni.',
          running: 'In uso: {{mhz}} MHz',
          rebootToApply: 'riavvia per applicare',
          rebootConfirm: 'Riavviare ora per applicare {{mhz}} MHz?'
        },
        videoClock: {
          title: 'Clock video',
          description: 'Profilo di clock degli encoder e dello scaler video',
          tip: 'Veloce fa funzionare il core del codec, lo scaler e il bus del codec a 594, 396 e 450 MHz; Standard ai valori di fabbrica del chip, 396, 300 e 360 MHz. Veloce riduce la latenza dello stream di circa 4 ms (H.265) e 3 ms (H.264), senza un aumento di temperatura misurabile. La modifica si applica al prossimo avvio del video.',
          standard: 'Standard',
          fast: 'Veloce',
          rates: 'Codec {{codec}} · scaler {{vpss}} · bus {{bus}} MHz',
          pending: 'si applica al prossimo avvio del video'
        },
        swap: {
          title: 'Scambia',
          disable: 'Disabilita',
          description: 'Imposta la dimensione del file di scambio',
          tip: 'Abilitare questa funzione potrebbe ridurre la durata utile della tua scheda SD!',
          active: 'Attivo - {{used}} di {{total}}',
          inactive: 'Configurato, ma non in uso'
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
          mode: 'Protocollo',
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
        usbIdentity: {
          title: 'Identità USB',
          description: 'Gli ID fornitore e prodotto e i nomi che l\'host remoto vede per questo dispositivo.',
          live: 'L\'host vede:',
          bootOverride: '/boot/usb.vid o /boot/usb.pid è impostato. Si applica finché è selezionato il predefinito.',
          vendorId: 'ID fornitore',
          productId: 'ID prodotto',
          manufacturer: 'Produttore',
          product: 'Nome prodotto',
          serial: 'Numero di serie',
          serialDesc: 'Da 12 a 64 lettere e cifre. Lasciare vuoto per mantenere quello della scheda.',
          serialPlaceholder: 'Quello della scheda',
          invalid: 'Controllare i campi evidenziati.',
          apply: 'Applica',
          restore: 'Ripristina predefinito',
          confirm: 'Cambiare l\'identità USB?',
          restoreConfirm: 'Ripristinare l\'identità USB predefinita?',
          warning: 'L\'host vede un nuovo dispositivo USB e lo enumera di nuovo: tastiera, mouse e disco virtuale si interrompono per alcuni secondi e l\'host potrebbe installare driver.',
          error: 'Impossibile leggere l\'identità USB.',
          presets: {
            default: 'Predefinito',
            logitech: 'Tastiera Logitech',
            microsoft: 'Tastiera Microsoft',
            dell: 'Tastiera Dell',
            custom: 'Personalizzato'
          }
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
        usbApply: {
          changed: 'Modificato',
          discard: 'Annulla',
          pending: 'Le modifiche non sono ancora state applicate.'
        },
        endpoints: {
          title: 'Slot USB',
          free: '{{free}} di {{total}} liberi',
          slots: 'Slot: {{count}}',
          full: "Slot USB liberi insufficienti. Disattiva prima qualcos'altro.",
          inactive:
            'Attivo, ma non in funzione: il controller USB ha esaurito gli slot. Disattiva un altro dispositivo e questo si avvia subito.',
          explain:
            'Il controller USB ha un numero fisso di slot (endpoint in ingresso), e tastiera e mouse ne occupano sempre alcuni. Se sono attivi più dispositivi di quanti ne entrino, tastiera e mouse restano e gli altri vengono disattivati.',
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
        ipv6: {
          title: 'IPv6',
          description: "Imposta come IronKVM ottiene l'indirizzo IPv6 sulla rete cablata",
          slaac: 'SLAAC',
          dhcpv6: 'DHCPv6',
          static: 'Manuale',
          off: 'Disattivato',
          address: 'Indirizzo IPv6',
          prefix: 'Lunghezza del prefisso',
          router: 'Router',
          apply: 'Applica',
          saved: 'Impostazioni IPv6 applicate e salvate',
          loadFailed: 'Impossibile leggere le impostazioni IPv6',
          saveFailed: 'Impossibile applicare le impostazioni IPv6',
          unsupported: 'Questo kernel non ha IPv6',
          noDhcpv6: 'DHCPv6 non è disponibile: questa immagine non ha un client DHCPv6',
          invalid:
            "Inserisci un indirizzo IPv6 valido, una lunghezza del prefisso da 16 a 126 e l'indirizzo del router (o lascialo vuoto)",
          offNote:
            "IPv6 è disattivato sull'interfaccia cablata. L'interfaccia web resta raggiungibile via IPv4.",
          keepsIpv4:
            "L'indirizzo IPv4 non cambia. Se l'applicazione fallisce, le impostazioni IPv6 precedenti vengono ripristinate.",
          useIpv4:
            'Questa pagina è aperta via IPv6. Apri IronKVM al suo indirizzo IPv4 per modificare le impostazioni IPv6.',
          current: 'Indirizzi IPv6 attuali',
          disabled: 'Disattivato',
          noAddress: 'Ancora nessun indirizzo IPv6',
          scopeLink: 'link-local',
          scopeUla: 'privato',
          scopeGlobal: 'globale'
        },
        webrtc: {
          title: 'WebRTC tramite port forwarding',
          description:
            'Permette agli spettatori fuori dalla tua rete di raggiungere il video tramite un port forwarding del router.',
          warning:
            "Inoltrare una porta espone questo KVM a Internet. Una VPN come Tailscale o NetBird resta il modo consigliato per accedere. Inoltra solo la porta UDP qui sotto, mai la porta dell'interfaccia web.",
          udpPort: 'Porta UDP fissa',
          udpPortHint:
            'Tutti gli spettatori condividono questa porta. Lascia vuoto per disattivarla. Inoltrala sul router a questo dispositivo.',
          publicAddress: 'Indirizzo pubblico',
          publicAddressHint:
            "L'indirizzo IP o il nome DNS dinamico a cui risponde il router. Richiede la porta UDP.",
          apply: 'Applica',
          saved: 'Salvato. I nuovi spettatori usano le nuove impostazioni.',
          loadFailed: 'Impossibile leggere le impostazioni WebRTC',
          saveFailed: 'Impossibile salvare le impostazioni WebRTC',
          invalidPort: 'La porta deve essere un numero da 1024 a 65535',
          needsPort: 'Un indirizzo pubblico richiede una porta UDP',
          active: 'Gli spettatori condividono la porta UDP {{port}}',
          inactive: 'Ogni spettatore usa una porta propria',
          offered: 'Offerto agli spettatori: {{addresses}}',
          cost:
            'Su una porta condivisa la scheda invia un pacchetto alla volta, il che può aumentare il carico della CPU con più spettatori e frame rate alti.'
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
        },
        syslog: {
          title: 'Log remoto',
          description:
            'Invia i log di sistema, del kernel e di IronKVM a un collettore syslog via UDP.',
          placeholder: 'host o host:porta, UDP, porta predefinita 514',
          save: 'Salva',
          turnOff: 'Disattiva',
          test: 'Invia messaggio di prova',
          sent: 'Inviato <v>{{message}}</v>. Cercalo sul collettore.',
          forwarding: 'Inoltro a <v>{{target}}</v>',
          local: 'Solo locale',
          savedButLocal:
            'Salvato <v>{{target}}</v>, ma il logger in esecuzione tiene i log in locale',
          savedButForwarding:
            'Salvato <v>{{target}}</v>, ma il logger in esecuzione inoltra a <v>{{active}}</v>',
          offButForwarding:
            'Disattivato, ma il logger in esecuzione inoltra ancora a <v>{{active}}</v>',
          unsupported: "Il logger di questa immagine ignora l'impostazione; aggiorna per usarla",
          loadFailed: "Impossibile leggere l'impostazione dei log",
          saveFailed: "Impossibile salvare l'impostazione dei log",
          testFailed: 'Impossibile inviare il messaggio di prova',
          metrics: 'Metriche',
          metricsDesc:
            "Prometheus può leggere l'URL qui sotto con un token bearer da <link>Chiavi API</link>.",
          metricsUrl: 'URL delle metriche',
          errors: {
            empty: 'Inserisci un host o host:porta',
            long: 'Troppo lungo per un nome host',
            ipv6: "Metti l'indirizzo IPv6 tra parentesi quadre, come [fd00::1]:514",
            brackets: "L'indirizzo tra parentesi non è un indirizzo IPv6",
            host: 'Usa solo lettere, cifre, punti e trattini, senza iniziare con punto o trattino',
            port: 'La porta deve essere un numero da 1 a 65535'
          }
        },
        allowlist: {
          title: 'Reti consentite',
          description:
            "Indirizzi che possono raggiungere l'interfaccia web, l'API, VNC e IPMI. Un indirizzo o CIDR per riga, ad esempio 192.168.1.0/24. Lasciare vuoto per consentire qualsiasi indirizzo.",
          everyone: 'Qualsiasi indirizzo può connettersi.',
          restricted:
            'Possono connettersi solo le reti elencate. Il dispositivo stesso e il collegamento di rete USB sono sempre consentiti.',
          placeholder: '192.168.1.0/24',
          yourAddress: 'Il tuo indirizzo: {{client}}',
          addMine: 'Aggiungi il mio indirizzo',
          save: 'Salva',
          allowAll: 'Consenti qualsiasi indirizzo',
          loadFailed: 'Impossibile caricare le reti consentite',
          saveFailed: 'Impossibile salvare le reti consentite'
        }
      },
      vpn: {
        kvmUrl: 'Indirizzo del KVM',
        moreTip: 'Altre azioni',
        updateTip: 'Aggiorna a {{version}}',
        loading: 'Caricamento...',
        okBtn: 'Sì',
        cancelBtn: 'No',
        restart: 'Riavviare {{name}}?',
        update: 'Aggiornare {{name}} alla versione {{version}}?',
        updateDesc: "Il demone si riavvia se è in esecuzione. L'accesso viene mantenuto.",
        notInstall: '{{name}} non è installato.',
        install: 'Installa',
        installing: 'Installazione',
        installFailed: 'Installazione non riuscita',
        retry: 'Riprova',
        connected: 'Connesso',
        connectedDesc:
          'Attivo, entra nella rete {{name}}. Disattivo, disconnette e arresta {{name}} per liberare memoria.',
        connectAtBoot: "Connetti all'avvio",
        connectAtBootDesc: "Entra nella rete {{name}} all'avvio del KVM.",
        restartService: 'Riavvia il servizio',
        needsLogin: 'Accesso richiesto',
        error: 'Errore',
        thisDevice: 'Questo dispositivo',
        memoryOf: '{{used}} (limite dei componenti aggiuntivi {{limit}})',
        memoryPressed:
          'Il gruppo di memoria dei componenti aggiuntivi è vicino al limite. Attivare lo swap può aiutare.',
        peersSummary: 'Peer: {{online}} online su {{total}}',
        peersSummaryIdle: 'Peer: {{online}} online, {{idle}} su richiesta, di {{total}}',
        showOffline: 'Mostra offline ({{offline}})',
        hideOffline: 'Nascondi offline',
        blocked:
          "{{other}} è attivo o si connette all'avvio. Può funzionare una sola VPN alla volta: disattiva prima Connesso e Connetti all'avvio per {{other}}.",
        deviceName: 'Nome dispositivo',
        deviceIP: 'IP dispositivo',
        account: 'Account',
        version: 'Versione',
        uptime: 'Tempo di attività',
        noPeers: 'Ancora nessun peer.',
        online: 'Online',
        offline: 'Offline',
        idle: 'Su richiesta: NetBird si connette quando il traffico lo richiede',
        memory: 'Memoria',
        uninstall: 'Disinstalla {{name}}',
        uninstallDesc: "Sei sicuro di voler disinstallare {{name}}? L'accesso resta sulla scheda.",
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
        logout: 'Esci',
        logoutDesc: 'Sei sicuro di voler uscire?',
        manualIntro: 'Oppure installalo a mano via SSH:',
        copyBinaries: "Copia tailscale e tailscaled in {{dir}} sull'IronKVM",
        linksFile: 'Nella stessa directory, crea un file chiamato links con queste due righe:',
        rebootRefresh: "Riavvia l'IronKVM, poi aggiorna questa pagina",
        loginServer: 'Server di controllo',
        loginServerDesc:
          "Lascia vuoto per usare Tailscale. Inserisci l'URL https di un server Headscale per usarlo al suo posto. Una modifica avvia un nuovo accesso, e il server precedente resta se il nuovo non risponde.",
        loginServerDefault: 'Tailscale (predefinito)',
        loginServerInvalid: 'Inserisci un URL https, ad esempio https://headscale.example.com',
        loginServerLocked: 'Esci per cambiare il server di controllo.',
        loginServerSaved: 'Server di controllo salvato'
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
        releaseNotes: 'Note di rilascio',
        ui: {
          title: 'Interfaccia web',
          desc: 'Installa o ripristina l\'interfaccia web senza aggiornare il server',
          served: 'In uso: {{version}}, API del server {{api}}',
          bundled: 'l\'interfaccia inclusa nell\'applicazione',
          bundledVersion: '{{version}} (inclusa nell\'applicazione)',
          checkOff: 'La verifica della firma è disattivata (/data/ironkvm/no-signature-check): qualsiasi pacchetto viene installato.',
          noKeys: 'Il portachiavi è vuoto: i pacchetti vengono installati senza verifica della firma.',
          check: 'Verifica',
          noPackage: 'Il server di aggiornamento non offre alcun pacchetto dell\'interfaccia web.',
          isCurrent: 'L\'interfaccia web {{version}} è già in uso.',
          incompatible: 'L\'interfaccia web {{version}} richiede l\'API del server {{api}}, questo server usa {{server}}. Aggiornare prima l\'applicazione.',
          installTo: 'Installa l\'interfaccia web {{version}}',
          installDesc: 'Il dispositivo verifica la firma, passa alla nuova interfaccia e conserva quella attuale per un ripristino. Poi questa pagina si ricarica.',
          chooseFiles: 'Scegli pacchetto',
          chooseHint: 'ironkvm-ui_<version>.tar.gz e il suo .minisig',
          noSignature: 'nessuna firma',
          pick: {
            noPackage: 'Scegliere un pacchetto ironkvm-ui_<version>.tar.gz.',
            tooMany: 'Scegliere un pacchetto e al massimo una firma.',
            badName: 'Non è un pacchetto dell\'interfaccia web (ironkvm-ui_<version>.tar.gz).',
            signatureMismatch: 'La firma appartiene a un altro pacchetto.'
          },
          install: 'Installa',
          installed: 'Interfaccia web installata. Ricaricamento...',
          rolledBack: 'Interfaccia web ripristinata. Ricaricamento...',
          rollback: 'Ripristina',
          rollbackDesc: 'Torna a {{version}}',
          rollbackConfirm: 'Ripristinare l\'interfaccia web {{version}}?'
        }
      },
      account: {
        title: 'Account',
        webAccount: 'Nome account web',
        role: 'Ruolo',
        roles: { admin: 'Amministratore', user: 'Utente', viewer: 'Spettatore' },
        shareLinks: {
          title: 'Link di condivisione',
          desc: 'Un link di condivisione mostra lo schermo a chiunque lo apra. Non dà accesso a tastiera, mouse, alimentazione, supporti o impostazioni e smette di funzionare alla scadenza o se viene revocato.',
          create: 'Crea link',
          label: 'Etichetta (facoltativa)',
          duration: 'Valido per',
          durations: {
            h1: '1 ora',
            h8: '8 ore',
            d1: '1 giorno',
            d7: '7 giorni',
            d30: '30 giorni'
          },
          empty: 'Nessun link di condivisione',
          expires: 'Scade il {{time}}',
          revoke: 'Revoca',
          revokeConfirm: 'Revocare questo link? Chi guarda tramite il link viene disconnesso.',
          revoked: 'Link revocato',
          loadFailed: 'Impossibile caricare i link di condivisione',
          createFailed: 'Impossibile creare il link di condivisione',
          revokeFailed: 'Impossibile revocare il link di condivisione',
          createdTitle: 'Link di condivisione creato',
          createdDesc: 'Copia il link adesso. Non potrà essere mostrato di nuovo.'
        },
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
      stop: 'Interrompi',
      cancel: 'Annulla',
      status: {
        connecting: 'Connessione al gateway...',
        connected: 'Sessione PicoClaw connessa',
        disconnected: 'Sessione PicoClaw chiusa',
        stopped: 'Richiesta di interruzione inviata',
        stopFailed: 'Impossibile inviare la richiesta di interruzione: la chat non è connessa.',
        stopTimeout:
          "PicoClaw non ha confermato l'interruzione. L'attività potrebbe essere ancora in corso.",
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
        name: 'PicoClaw (build di IronKVM)',
        update: 'Aggiorna a {{version}}',
        installing: 'Installazione PicoClaw',
        success: 'PicoClaw installato correttamente',
        failed: 'Impossibile installare PicoClaw',
        uninstalling: 'Disinstallazione del runtime in corso...',
        uninstalled: 'Runtime disinstallato correttamente.',
        uninstallFailed: 'Disinstallazione non riuscita.',
        requiredTitle: 'PicoClaw non è installato',
        requiredDescription: 'Installa PicoClaw prima di avviare il runtime PicoClaw.',
        progressDescription: 'PicoClaw è in fase di installazione.',
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
        apiKeyOptionalPlaceholder: 'Non necessaria per ollama, lmstudio o vllm',
        save: 'Salva',
        saving: 'Salvataggio',
        saved: 'Configurazione del modello salvata',
        saveFailed: 'Impossibile salvare la configurazione del modello',
        invalid: 'Identificatore del modello, API Base URL e chiave API sono obbligatori',
        invalidNoKey: 'Identificatore del modello e API Base URL sono obbligatori'
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
    speaker: {
      title: 'Altoparlante',
      unmute: 'Riattiva audio',
      mute: 'Disattiva audio',
      hostIdle: "L'host non invia audio",
      hostIdleHint: "Riproduci qualcosa sull'host o scegli il KVM come uscita audio.",
      room: 'Microfono della stanza',
      roomListen: 'Ascolta la stanza',
      roomVolume: 'Volume',
      roomHint: 'Chiunque guardi vede quando è acceso.',
      roomLive: 'Microfono attivo',
      roomLiveBy: 'Microfono della stanza acceso: {{names}}'
    },
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
      mediaAdd: 'Aggiungi immagine',
      mediaMoreOptions: 'Altre opzioni',
      mediaSettings: 'Impostazioni supporti',
      collapse: 'Comprimi menu',
      expand: 'Espandi il menu',
      more: 'Altro',
      media: 'Supporti',
      tools: 'Strumenti',
      text: 'Testo',
      advanced: 'Avanzate',
      mediaMounted: 'Montato',
      mediaLibrary: 'Libreria',
      textToHost: "Verso l'host",
      textFromHost: "Dall'host"
    },
    serverContract: {
      incompatibleTitle: 'Questa interfaccia web non supporta questo server',
      incompatible: 'L\'interfaccia web supporta l\'API del server {{ui}}, il server usa l\'API {{server}}. Alcune funzioni potrebbero non funzionare. Aggiornare l\'applicazione o ripristinare l\'interfaccia web in Impostazioni > Aggiornamento.',
      dismiss: 'Chiudi'
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
    },
    failsafe: {
      title: 'Video disattivato: il server è avviato in modalità provvisoria',
      body: "La libreria video ha fatto arrestare il server più volte di seguito, quindi ora il server funziona senza video. Tutto il resto funziona: impostazioni, aggiornamenti, log, controllo dell'alimentazione e terminale. Riprova il video dopo un aggiornamento o una correzione. Anche il riavvio di IronKVM riprova il video.",
      showLog: 'Mostra log',
      hideLog: 'Nascondi log',
      retry: 'Riprova video',
      retrying: 'Riavvio in corso...'
    },
    totp: {
      login: {
        title: 'Autenticazione a due fattori',
        description:
          "Inserisci il codice a 6 cifre dell'app di autenticazione, oppure uno dei tuoi codici di recupero.",
        placeholder: 'Codice o codice di recupero',
        noEmptyCode: 'Inserisci un codice',
        verify: 'Verifica',
        back: 'Torna alla password',
        invalidCode: 'Codice non valido',
        locked: 'Troppi codici errati. Riprova tra qualche minuto.',
        expired: "L'accesso è scaduto. Inserisci di nuovo la password."
      },
      settings: {
        title: 'Autenticazione a due fattori',
        description: "Richiedi un codice da un'app di autenticazione dopo la password.",
        scope:
          "Vale solo per l'accesso web. VNC, IPMI e Redfish usano ancora solo la password, e le chiavi API funzionano senza codice.",
        on: 'Attiva',
        off: 'Disattiva',
        enable: 'Attiva',
        disable: 'Disattiva',
        regenerate: 'Nuovi codici di recupero',
        codesLeft: 'Codici di recupero rimasti: {{count}}',
        enableTitle: "Attiva l'autenticazione a due fattori",
        disableTitle: "Disattiva l'autenticazione a due fattori",
        regenerateTitle: 'Nuovi codici di recupero',
        confirmPassword: 'Conferma con la tua password',
        continue: 'Continua',
        scan: "Scansiona il codice QR con un'app di autenticazione, oppure inserisci il segreto a mano. Poi digita il codice mostrato dall'app.",
        secret: 'Segreto',
        enterCode: "Codice dell'app",
        verifyAndEnable: 'Verifica e attiva',
        codesOnce:
          'Salva ora questi codici di recupero. Ognuno vale una sola volta al posto di un codice, e non verranno più mostrati.',
        recoveryCodes: 'Codici di recupero IronKVM',
        download: 'Scarica',
        saved: 'Li ho salvati',
        disabled: 'Autenticazione a due fattori disattivata',
        wrongPassword: 'La password non è corretta',
        setupExpired: 'La configurazione è scaduta. Ricomincia.',
        loadFailed: "Impossibile caricare lo stato dell'autenticazione a due fattori",
        failed: 'La modifica non è riuscita'
      },
      users: {
        tag: '2FA',
        reset: 'Reimposta 2FA',
        resetConfirm: "Disattivare l'autenticazione a due fattori per {{username}}?",
        resetDone: 'Autenticazione a due fattori reimpostata'
      }
    },
    tlsCert: {
      title: 'Certificato',
      description: 'Il certificato presentato dal server HTTPS.',
      uploaded: 'Caricato',
      selfSigned: 'Autofirmato',
      httpsOff: 'HTTPS è disattivato. Il certificato verrà usato quando HTTPS sarà attivato.',
      none: 'Nessun certificato. Ne viene generato uno quando si attiva HTTPS.',
      subject: 'Soggetto',
      issuer: 'Emittente',
      names: 'Nomi',
      validFrom: 'Valido dal',
      validUntil: 'Valido fino al',
      keyType: 'Chiave',
      chain: 'Certificati nella catena',
      fingerprint: 'SHA-256',
      restore: 'Ripristina autofirmato',
      restoreConfirm: 'Tornare al certificato autofirmato?',
      restoreDesc:
        'Il certificato e la chiave caricati verranno eliminati. I browser mostreranno di nuovo un avviso per il certificato autofirmato.',
      restored: 'Certificato autofirmato ripristinato',
      restoreFailed: 'Impossibile ripristinare il certificato autofirmato',
      uploadTitle: 'Carica un certificato',
      uploadDescription:
        'Formato PEM. Metti prima il certificato del server, poi gli eventuali intermedi. La chiave non deve essere cifrata.',
      certField: 'Catena di certificati',
      keyField: 'Chiave privata',
      chooseFile: 'Scegli file',
      readFailed: 'Impossibile leggere il file',
      check: 'Controlla',
      install: 'Installa',
      installConfirm: 'Installare questo certificato?',
      installDesc:
        'Le nuove connessioni lo usano subito. Le pagine aperte continuano a funzionare; ricarica per vedere il nuovo certificato.',
      installed: 'Certificato installato',
      installFailed: 'Impossibile installare il certificato',
      checkFailed: 'Il certificato è stato rifiutato',
      loadFailed: 'Impossibile caricare il certificato',
      warnings: {
        expiresSoon: 'Il certificato scade entro 30 giorni.',
        nameMismatch:
          'Il certificato non copre: {{names}}. I browser mostreranno un avviso quando il dispositivo viene aperto con questi nomi.',
        selfSigned:
          'Questo certificato è autofirmato. I browser lo considerano attendibile solo dopo averlo installato su ogni computer.'
      },
      errors: {
        tooLarge: 'Il file è troppo grande.',
        noCertificate: 'Nessun certificato PEM trovato.',
        badCertificate:
          'Impossibile leggere il certificato. Incolla un testo PEM che inizi con -----BEGIN CERTIFICATE-----.',
        noKey: 'Nessuna chiave privata PEM trovata.',
        encryptedKey: 'La chiave privata è cifrata. Rimuovi prima la passphrase.',
        badKey: 'Impossibile leggere la chiave privata.',
        keyMismatch: 'La chiave privata non appartiene al certificato.',
        expired: 'Il certificato, o uno della sua catena, è scaduto.',
        notYetValid: "Il certificato non è ancora valido. Controlla l'orologio del dispositivo.",
        badChain:
          'La catena è in ordine errato o incompleta. Metti prima il certificato del server, poi ogni emittente.',
        weakKey: 'La chiave è troppo corta. Usa RSA 2048 o superiore, oppure ECDSA.',
        notServerCert: 'Il certificato non è per un server TLS.'
      }
    }
  }
};

export default it;
