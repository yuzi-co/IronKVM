const nb = {
  translation: {
    feedback: {
      enabled: '{{name}} slått på',
      disabled: '{{name}} slått av',
      failed: 'Forespørselen mislyktes. Prøv igjen.',
      network: 'Fikk ikke kontakt med enheten. Sjekk tilkoblingen og prøv igjen.',
      saved: 'Lagret',
      timeout: 'Enheten brukte for lang tid på å svare. Prøv igjen.'
    },
    common: {
      copy: 'Kopier',
      copied: 'Kopiert',
      copyFailed: 'Kunne ikke kopiere. Merk teksten og kopier den manuelt.',
      notUpdating: 'Oppdateres ikke: siste oppdatering mislyktes.',
      off: 'Av',
      running: 'Kjører',
      save: 'Lagre',
      cancel: 'Avbryt',
      delete: 'Slett',
      remove: 'Fjern'
    },
    head: {
      desktop: 'Eksternt skrivebord',
      login: 'Logg inn',
      changePassword: 'Endre passord',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Passordet er endret. Logg inn med det nye passordet.',
      cookieRejected:
        'Nettleseren nektet å lagre økten. En informasjonskapsel fra en tidligere HTTPS-økt kan ikke erstattes over vanlig http. Slett informasjonskapslene for denne adressen, eller åpne et privat vindu, og logg inn på nytt.',
      login: 'Logg inn',
      placeholderUsername: 'Brukernavn',
      placeholderPassword: 'Passord',
      placeholderCurrentPassword: 'Nåværende passord',
      placeholderPassword2: 'Oppgi passord igjen',
      noEmptyUsername: 'Brukernavn påkrevd',
      noEmptyPassword: 'Passord påkrevd',
      passwordLength: 'Passordet må være mellom 8 og 72 tegn',
      noAccount:
        'Kunne ikke hente brukerinformasjon. Vennligst last inn siden på nytt eller gjenopprett passord',
      invalidUser: 'Ugyldig brukernavn eller passord',
      locked: 'For mange pålogginger, vennligst prøv igjen senere',
      globalLocked: 'System under beskyttelse, prøv igjen senere',
      error: 'Uventet feil',
      invalidCurrentPassword: 'Nåværende passord er feil',
      changePassword: 'Endre passord',
      changePasswordDesc:
        'For sikkerheten til enheten, vennligst endre passordet ditt for web-innlogging.',
      differentPassword: 'Passordene er ikke like',
      illegalUsername: 'Brukernavn inneholder tegn som ikke er tillat',
      illegalPassword: 'Passord inneholder tegn som ikke er tillat',
      forgetPassword: 'Glemt passord',
      ok: 'Ok',
      cancel: 'Avbryt',
      loginButtonText: 'Logg inn',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the IronKVM for 10 seconds.',
        reset3: 'Standard webkonto:',
        reset4: 'Standard SSH-konto:',
        change1: 'Merk at denne handlingen endrer følgende passord:',
        change2: 'Passord for webinnlogging',
        change3: 'Systemets root-passord (SSH-innloggingspassord)',
        change4: 'For å tilbakestille passordene holder du BOOT-knappen på IronKVM inne.',
        resetDocs: 'Se maskinvaredokumentasjonen for detaljerte trinn:',
        hardwareDocs: 'Sipeed NanoKVM-wiki'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Konfigurer Wi-Fi for IronKVM',
      success: 'Please check the network status of IronKVM and visit the new IP address.',
      failed: 'Operasjonen mislyktes, prøv igjen.',
      invalidMode:
        'Gjeldende modus støtter ikke nettverksoppsett. Gå til enheten din og aktiver Wi-Fi konfigurasjonsmodus.',
      confirmBtn: 'Ok',
      finishBtn: 'Ferdig',
      ap: {
        authTitle: 'Autentisering kreves',
        authDescription: 'Vennligst skriv inn AP passordet for å fortsette',
        authFailed: 'Ugyldig AP passord',
        passPlaceholder: 'AP passord',
        verifyBtn: 'Bekreft'
      },
      ssidRequired: 'Skriv inn nettverksnavnet, opptil 32 tegn',
      passwordLength: 'Passordet er 8 til 63 tegn. La det stå tomt for et åpent nettverk.',
      passwordOptional: 'Passord (tomt for et åpent nettverk)',
      lost: 'Kortet sluttet å svare. Det kan ha koblet seg til nettverket og lukket oppsettshotspoten. Hvis hotspoten kommer tilbake, mislyktes tilkoblingen: koble til den igjen og prøv på nytt.',
      done: 'Oppsettet er ferdig. Koble denne enheten til ditt vanlige nettverk igjen og åpne kortet på den nye adressen.'
    },
    screen: {
      viewOnly: 'Bare visning',
      viewOnlyTip:
        'Denne fanen slutter å sende tastatur og mus til verten. Skript, mus-jiggleren og andre seere påvirkes ikke.',
      viewOnlyOff: 'Slå av bare visning',
      viewOnlyBlocked: 'Bare visning er på, så ingenting ble sendt til verten',
      pauseHidden: 'Pause når fanen er skjult',
      pauseHiddenTip:
        'Stopper video og lyd noen sekunder etter at fanen skjules, og starter dem igjen når du kommer tilbake.',
      screenshot: 'Skjermbilde',
      screenshotTip: 'Lagrer vertens skjerm som PNG i full opptaksstørrelse.',
      screenshotFailed: 'Skjermbilde mislyktes',
      stream: {
        ok: 'bilde OK',
        noSignal: 'ingen signal',
        failed: 'strømmen feilet'
      },
      codecNoWebrtcHevc: 'Denne nettleseren kan ikke motta H.265 over WebRTC',
      codecNoHevc: 'Denne nettleseren kan ikke dekode H.265',
      codecNote:
        'Kortet har én koder, så dette endrer strømmen for alle seere. Koble til på nytt for å bruke det i en pågående WebRTC-økt.',
      codec: 'Kodek',
      updateFailed: 'Innstillingen ble ikke brukt',
      scale: 'Skala',
      title: 'Skjerm',
      video: 'Video-kodek',
      videoDirectTips: 'Aktiver HTTPS i "Innstillinger > Enhet" for å bruke denne modusen',
      resolution: 'Oppløsning',
      ocr: {
        title: 'Les tekst (OCR)',
        tips: 'Teksten gjenkjennes i denne nettleseren. Du kan rette den før du kopierer den.',
        hint: 'Dra over teksten som skal leses. Trykk på Esc for å avbryte.',
        noPicture: 'Vent på videoen, og dra deretter over teksten som skal leses.',
        cancel: 'Avbryt',
        language: 'Språk',
        languages: {
          eng: 'Engelsk'
        },
        preview: 'Valgt område',
        capturing: 'Tar opp skjermen...',
        loading: 'Laster inn tekstgjenkjenning...',
        recognizing: 'Leser teksten...',
        noText: 'Fant ingen tekst i det valgte området.',
        copy: 'Kopier',
        copied: 'Kopiert til utklippstavlen',
        copyFailed: 'Kunne ikke kopiere til utklippstavlen',
        selectAgain: 'Velg på nytt',
        unsupported:
          'Denne nettleseren kan ikke kjøre tekstgjenkjenning. Den krever WebAssembly SIMD, som dagens nettlesere har.',
        captureFailed: 'Kunne ikke ta opp skjermen.',
        outside: 'Det valgte området er utenfor bildet.',
        recognizeFailed: 'Tekstgjenkjenningen mislyktes.'
      },
      controlRegion: {
        title: 'Musekalibrering',
        description:
          'Bruk denne innstillingen når den kontrollerte enheten bruker en oppløsning som ikke er 16:9, og markøren er forskjøvet vannrett eller loddrett.',
        off: 'Av',
        auto: 'Automatisk',
        autoWarning:
          'Kalibreringen kan mislykkes hvis brukerprogrammet har en helt svart bakgrunn.',
        manual: 'Manuell',
        selectedResolution: 'Oppløsning for valgt område',
        unused: 'Ikke i bruk',
        originalResolution: 'Opprinnelig oppløsning',
        selectResolution: 'Velg opprinnelig oppløsning',
        addResolution: 'Legg til egendefinert oppløsning',
        add: 'Legg til',
        duplicateResolution: 'Denne oppløsningen finnes allerede.',
        width: 'Bredde',
        height: 'Høyde',
        apply: 'Beregn og bruk',
        invalidResolution: 'Angi en gyldig opprinnelig oppløsning når videoen er klar.',
        select: 'Velg område',
        clear: 'Gjenopprett automatisk registrering',
        saveFailed: 'Kunne ikke lagre inndataområdet.',
        tooSmall: 'Det valgte området er for lite.',
        previewUnavailable: 'Forhåndsvisning er utilgjengelig',
        clearConfirm: 'Gjenopprette automatisk registrering av svarte kanter?',
        dragHint: 'Dra for å velge området på det eksterne skrivebordet',
        finish: 'Ferdig',
        confirm: 'Bekreft',
        cancel: 'Avbryt'
      },
      auto: 'Automatisk',
      autoTips:
        'Skjermriving eller peker-forskyvning kan oppstå ved enkelte oppløsninger. Prøv å justere den eksterne vertens oppløsning eller skru av automatisk modus.',
      fps: 'FPS',
      customizeFps: 'Tilpass',
      quality: 'Kvalitet',
      qualityLossless: 'Best',
      qualityHigh: 'Høy',
      qualityMedium: 'Middels',
      qualityLow: 'Lav',
      frameDetect: 'Bildefrekvensoppdagelse',
      frameDetectTip:
        'Kalkuler forskjellen mellom bilder. Stopper overføring av video når det ikke oppdages forskjell på den eksterne vertens skjerm.',
      resetHdmi: 'Tilbakestill HDMI',
      mixedH264: {
        title: 'H.264-strømmekonflikt',
        description:
          'H.264 Direct og H.264 WebRTC brukes samtidig. Dette kan føre til skjermriving eller ødelagt video. Bruk bare én H.264-modus.'
      },
      webrtcConnectionFailed: {
        title: 'WebRTC-tilkobling mislyktes',
        description: 'Kontroller nettverkstilkoblingen eller bytt videomodus.'
      },
      captureStatus: {
        hdmiError: 'HDMI-skjermfeil',
        unsupportedResolution: 'Gjeldende oppløsning støttes ikke',
        retrieving: 'Henter skjermbilde...',
        changingResolution: 'Bytter oppløsning...',
        updateFailed: 'Skjermen kan ikke oppdateres akkurat nå',
        videoError: 'Feil ved videovisning',
        noHdmi: 'Ingen HDMI-signal oppdaget',
        unavailable: 'Skjermen kan ikke vises akkurat nå'
      },
      directConnectionFailed: 'Tilkoblingen til videostrømmen mislyktes'
    },
    keyboard: {
      close: 'Lukk',
      title: 'Åpne tastatur',
      paste: 'Lim inn',
      tips: 'Skriver teksten på verten som tastetrykk. Velg tastaturoppsettet verten bruker.',
      placeholder: 'Vennligst angi teksten du vil lime inn',
      submit: 'Lim inn',
      virtual: 'Åpne tastatur',
      readClipboard: 'Les fra utklippstavlen',
      clipboardPermissionDenied:
        'Utklippstavle tillatelse nektet. Tillat utklippstavletilgang i nettleseren din.',
      clipboardReadError: 'Kunne ikke lese utklippstavlen',
      mediaKeys: {
        title: 'Medietaster',
        mute: 'Demp',
        volumeDown: 'Lavere volum',
        volumeUp: 'Høyere volum',
        previous: 'Forrige spor',
        playPause: 'Spill av eller pause',
        next: 'Neste spor',
        stop: 'Stopp'
      },
      pasting: {
        layout: 'Tastaturoppsett på verten',
        layouts: {
          us: 'Engelsk (USA)',
          uk: 'Engelsk (Storbritannia)',
          de: 'Tysk',
          fr: 'Fransk',
          es: 'Spansk',
          it: 'Italiensk',
          ptBr: 'Portugisisk (Brasil)',
          se: 'Svensk / finsk',
          ru: 'Russisk',
          ja: 'Japansk',
          ko: 'Koreansk'
        },
        speed: 'Skrivehastighet',
        speeds: {
          fast: 'Rask',
          normal: 'Normal',
          slow: 'Sakte'
        },
        estimate: 'Skrivetid: omtrent {{duration}}',
        untypeable: 'Tegn dette oppsettet ikke kan skrive: {{count}}',
        untypeableAt: 'linje {{line}}, kolonne {{column}}',
        skipUntypeable: 'Skriv resten',
        shortcut: '{{shortcut}} skriver utklippstavlen på verten med en gang.',
        clipboardUnavailable:
          'Nettleseren lar bare en side lese utklippstavlen over HTTPS. Lim inn teksten i feltet med Ctrl+V.',
        clipboardEmpty: 'Utklippstavlen inneholder ingen tekst.',
        tooLong: 'Teksten er for lang. Grensen er {{max}} tegn.',
        inProgress: 'En innliming skrives allerede.',
        typing: 'Skriver på verten',
        done: 'Tekst skrevet',
        canceled: 'Innliming avbrutt',
        failed: 'Innliming mislyktes',
        cancel: 'Avbryt',
        controlBusy: 'En annen kontroller bruker tastaturet.',
        hidError: 'Tastetrykkene kunne ikke sendes til verten.'
      },
      shortcut: {
        sendFailed: 'Ikke sendt: inndatatilkoblingen er nede',
        title: 'Snarveier',
        custom: 'Egendefinert',
        capture: 'Klikk her for å ta en snarvei',
        clear: 'Tøm',
        save: 'Lagre',
        captureTips:
          'Registrering av systemtaster (som Windows-tasten) krever fullskjermtillatelse.',
        enterFullScreen: 'Veksle fullskjermmodus.'
      },
      leaderKey: {
        saveFailed: 'Kunne ikke lagre ledertasten',
        title: 'Leader-tast',
        desc: 'Omgå nettleserrestriksjoner og send systemsnarveier direkte til den eksterne verten.',
        howToUse: 'Hvordan bruke',
        simultaneous: {
          title: 'Samtidig modus',
          desc1: 'Hold Leader-tasten inne, og trykk deretter på snarveien.',
          desc2: 'Intuitivt, men kan komme i konflikt med systemsnarveier.'
        },
        sequential: {
          title: 'Sekvensiell modus',
          desc1:
            'Trykk på Leader-tasten → trykk på snarveien i rekkefølge → trykk på Leader-tasten igjen.',
          desc2: 'Krever flere trinn, men unngår fullstendig systemkonflikter.'
        },
        enable: 'Aktiver Leader-tast',
        tip: 'Når denne tasten er tilordnet som Leader-tast, fungerer den bare som snarveisutløser og mister standardoppførselen.',
        placeholder: 'Trykk på Leader-tasten',
        shiftRight: 'Høyre Shift',
        ctrlRight: 'Høyre Ctrl',
        metaRight: 'Høyre Win',
        submit: 'Lim inn',
        recorder: {
          rec: 'REC',
          activate: 'Aktiver taster',
          input: 'Trykk snarveien...'
        }
      }
    },
    mouse: {
      jiggler: 'Musebeveger',
      keyJiggler: 'Tastetrykker',
      keyJigglerF15: 'F15-tast',
      keyJigglerShift: 'Shift-tast',
      keyJigglerCtrl: 'Ctrl-tast',
      keyJigglerF15Tip: 'F15 forstyrrer minst: ingen vanlige systemer eller apper bruker den',
      title: 'Mus',
      cursor: 'Markørstil',
      default: 'Vanlig',
      pointer: 'Hånd',
      cell: 'Celle',
      text: 'Tekst',
      grab: 'Grip',
      hide: 'Skjul',
      mode: 'Modus',
      absolute: 'Absolutt',
      relative: 'Relativ',
      absoluteShort: 'Absolutt',
      relativeShort: 'Relativ',
      touch: 'Berøring',
      touchShort: 'Berøring',
      absoluteStalled: 'Målmaskinen ignorerer den absolutte musen',
      absoluteStalledDesc:
        'Målmaskinen har sluttet å hente absolutte muserapporter, så pekerbevegelser går tapt. Tastaturet påvirkes ikke. Gjenoppretting av USB løser det ofte; relativ modus bruker et annet endepunkt.',
      useRelative: 'Bytt til relativ modus',
      direction: 'Rullehjulretning',
      scrollUp: 'Som på denne datamaskinen',
      scrollDown: 'Omvendt (naturlig rulling)',
      speed: 'Rullehjulhastighet',
      fast: 'Rask',
      slow: 'Sakte',
      requestPointer: 'Bruker relativ modus. Vennligsk klikk på skrivebordet for vise musepeker.',
      resetHid: 'Gjenopprett HID',
      hidOnly: {
        switchFailed: 'Kunne ikke bytte modus. Sjekk tilkoblingen og prøv igjen.',
        title: 'Kun HID-modus',
        desc: 'Hvis musen og tastaturet slutter å svare og tilbakestilling av HID ikke hjelper, kan det være et kompatibilitetsproblem mellom IronKVM og enheten. Prøv å aktivere HID-Only-modus for bedre kompatibilitet.',
        tip1: 'Aktivering av HID-Only-modus vil demontere den virtuelle U-disken og det virtuelle nettverket',
        tip2: 'I HID-Only-modus er bildemontering deaktivert',
        rebuild: 'Bytte av modus bygger opp USB-tilkoblingen på nytt. IronKVM starter ikke på nytt',
        enable: 'Aktiver HID-Only-modus',
        disable: 'Deaktiver HID-bare-modus'
      },
      resetHidDone: 'USB HID er tilbakestilt',
      resetHidFailed: 'Tilbakestilling av USB HID mislyktes'
    },
    image: {
      driveLoaded: 'avbildning satt inn',
      driveWarning: 'se advarslene',
      warning: {
        missing: 'Avbildningsfilen er slettet. Verten leser den gamle kopien til du løser den ut.',
        writable: 'Lese-skrive: verten kan endre denne avbildningen.',
        tooBigForCd: 'For stor for CD-stasjonen ({{size}}, grense {{max}}). Bruk disken.',
        tooSmallForCd: 'For liten for CD-stasjonen ({{size}}). Bruk disken.',
        empty: 'Filen er tom, trolig etter en mislykket opplasting eller nedlasting.'
      },
      delete: 'Slett',
      inUse: 'I bruk. Løs det ut før du sletter det.',
      retry: 'Prøv igjen',
      loadFailed: 'Kunne ikke laste listen over image',
      readOnlyLocked: 'Løs ut disken for å endre dette. Det gjelder når et image settes inn.',
      title: 'Bilder',
      loading: 'Laster...',
      empty: 'Ingen funnet',
      mountMode: 'Monteringsmodus',
      mountFailed: 'Montering feilet',
      mountDesc:
        'På noen systemer er det nødvendig å koble fra den virtuelle disken på den eksterne verten før man kan montere arkivfilen.',
      unmountFailed: 'Avmontering mislyktes',
      unmountDesc:
        'På noen systemer må du manuelt løse ut fra den eksterne verten før du demonterer bildet.',
      refresh: 'Oppdater bildelisten',
      disk: 'Disk',
      cdrom: 'CD',
      driveEmpty: 'Tom',
      eject: 'Løs ut',
      readOnly: 'Skrivebeskyttet',
      readOnlyTip: 'Gjelder for neste bilde som settes inn i disken.',
      noDrives: 'Ingen virtuelle stasjoner. Slå på virtuell disk i Innstillinger.',
      insertFailed: 'Innsetting mislyktes',
      ejectFailed: 'Utløsing mislyktes',
      insertInto: 'Settes inn i {{drive}}. Klikk for å endre.',
      loadedIn: 'Satt inn i {{drive}}',
      attention: 'Merknad',
      deleteConfirm: 'Er du sikker på at du vil slette dette bildet?',
      okBtn: 'Ja',
      cancelBtn: 'Nei',
      deleteFailed: 'Sletting mislyktes',
      ventoy: {
        statusNoKernel: 'Støttes ikke av denne fastvaren',
        statusNotInstalled: 'Ikke installert',
        statusReady: 'Klar',
        statusSelected: 'Valgte avbildninger: {{count}}',
        statusInDrive: 'I diskstasjonen, {{size}}',
        noKernel:
          'Kjernen i denne fastvaren har ikke støtte for device-mapper, så Ventoy kan ikke brukes før en avbildning med slik støtte er installert.',
        installDesc: 'Start verten fra flere avbildninger på én disk, uten å kopiere dem.',
        install: 'Installer',
        installing: 'Laster ned Ventoy, omtrent 20 MB. Dette kan ta noen minutter.',
        needsData: 'Ventoy trenger en IronKVM-avbildning med /data-partisjonen montert.',
        uninstall: 'Avinstaller',
        uninstallConfirm: 'Fjerne Ventoy-filene?',
        noImages: 'Ingen avbildninger å legge på Ventoy-disken.',
        onDisk: 'På Ventoy-disken',
        missing: 'Mangler: {{file}}',
        remove: 'Fjern fra Ventoy-disken',
        setHint:
          'Utvalget av avbildninger kan bare endres mens Ventoy-disken ikke er i en stasjon.',
        useAsDisk: 'Bruk som virtuell disk',
        failed: 'Ventoy-forespørselen mislyktes',
        secureBoot:
          'Med Secure Boot på må verten registrere Ventoys nøkkel i MokManager én gang. Nøkkelfilen ENROLL_THIS_KEY_IN_MOKMANAGER.cer ligger på VTOYEFI-partisjonen.',
        readOnly:
          'Verten ser disken som skrivebeskyttet, så Ventoy-persistens og ventoy.json på stasjonen fungerer ikke.'
      },
      tips: {
        title: 'Hvordan laste opp',
        usb1: 'Koble til IronKVM-enheten til din datamaskin med USB.',
        usb2: 'Sikre at den virtuelle disken er montert (Innstillinger - Virtuell disk).',
        usb3: 'Åpne den virtuelle disken på datamaskinen din og kopier arkivfilen til rot-mappen på den virtuelle disken.',
        scp1: 'Sikre at IronKVM-enheten og datamaskinen din er tilkoblet det samme lokale nettverket.',
        scp2: 'Åpne en terminal på datamaskinen din og bruk SCP-kommandoen til å laste opp arkivfilen til mappen /data på IronKVM-enheten.',
        scp3: 'Eksempel: scp sti-til-din-arkivfil root@din-nanokvm-ip:/data',
        tfCard: 'TF-kort',
        tf1: 'Denne metoden er støttet på datamskiner med Linux',
        tf2: 'Ta TF-kortet ut av IronKVM-enheten (hvis du har FULL-versjonen, demonter kabinettet først).',
        tf3: 'Sett inn TF-kortet i en kortleser og koble den til datamaskinen din.',
        tf4: 'Kopiér arkivfilen til mappen /data på TF-kortet.',
        tf5: 'Sett inn TF-kortet i IronKVM-enheten.'
      }
    },
    script: {
      title: 'Skript',
      upload: 'Last opp',
      run: 'Kjør',
      runBackground: 'Kjør i bakgrunnen',
      runFailed: 'Kjøring feilet',
      attention: 'Merknad',
      delDesc: 'Er du sikker på at du vil slette denne filen?',
      confirm: 'Ja',
      cancel: 'Nei',
      delete: 'Slett',
      close: 'Lukk',
      empty: 'Ingen skript ennå. Last opp en .sh- eller .py-fil for å kjøre den på kortet.',
      loadFailed: 'Kunne ikke laste inn skriptene',
      uploaded: 'Skript lastet opp',
      uploadFailed: 'Kunne ikke laste opp skriptet',
      started: 'Skriptet er startet i bakgrunnen',
      deleteFailed: 'Kunne ikke slette skriptet',
      waitLimit: 'Venter på at skriptet blir ferdig, i opptil {{minutes}} minutter.',
      timedOut:
        'Skriptet kjørte lenger enn {{minutes}} minutter, og siden sluttet å vente. Det kan fortsatt kjøre på kortet.'
    },
    terminal: {
      invalidBaud: 'Denne baudhastigheten støttes ikke.',
      invalidPort: 'Skriv inn en enhetsbane under /dev, for eksempel /dev/ttyS1.',
      invalidSettings: 'Ugyldige innstillinger for seriellport. Dette er kortets eget skall.',
      disconnected: 'Frakoblet. Trykk Enter for å koble til igjen.',
      title: 'Terminal',
      nanokvm: 'IronKVM',
      serial: 'Seriell port',
      serialPort: 'Seriell port',
      serialPortPlaceholder: 'Vennligst angi den serielle porten',
      baudrate: 'Baud-rate',
      parity: 'Paritet',
      parityNone: 'Ingen',
      parityEven: 'Lik',
      parityOdd: 'Ulik',
      flowControl: 'Strømningskontroll',
      flowControlNone: 'Ingen',
      flowControlSoft: 'Programvare',
      flowControlHard: 'Maskinvare',
      dataBits: 'Databiter',
      stopBits: 'Stoppbiter',
      confirm: 'Ok'
    },
    wol: {
      no: 'Nei',
      yes: 'Ja',
      deleteConfirm: 'Slette denne lagrede adressen?',
      delete: 'Slett',
      wake: 'Vekk',
      rename: 'Gi nytt navn',
      showMac: 'Vis MAC-adresse',
      showName: 'Vis navn',
      requestFailed: 'Fikk ikke kontakt med enheten for å sende kommandoen',
      deleteFailed: 'Kunne ikke slette',
      renameFailed: 'Kunne ikke gi nytt navn',
      title: 'Wake-on-LAN',
      sending: 'Sender kommando...',
      sent: 'Kommando sendt',
      input: 'Vennligst angi MAC-adressen',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'Opplastingen mislyktes',
      uploadSuccess: 'Opplasting fullført',
      uploading: 'Laster opp: {{file}}',
      downloadingPercent: 'Laster ned ({{percent}}): {{file}}',
      downloading: 'Laster ned: {{file}}',
      title: 'Bildedaster',
      input: 'Vennligst skriv inn et eksternt bilde URL',
      ok: 'Ok',
      disabled: '/data partisjonen er RO, så vi kan ikke laste ned bildet',
      uploadbox: 'Slipp filen her eller klikk for å velge',
      inputfile: 'Vennligst skriv inn bildefilen',
      NoISO: 'Ingen ISO',
      sha256: 'SHA-256 (valgfritt)',
      sha256Placeholder: 'Skriv inn en SHA-256-kontrollsum på 64 tegn',
      invalidSHA256: 'SHA-256 må være en heksadesimal streng på 64 tegn',
      failed: 'Nedlasting mislyktes',
      success: 'Nedlasting fullført',
      checksumFailed: 'Nedlasting mislyktes: SHA-256-verifisering mislyktes',
      cancel: 'Avbryt',
      cancelFailed: 'Kunne ikke avbryte nedlastingen',
      bootMenu: 'Oppstartsmeny (netboot.xyz)',
      bootMenuPresent: '{{file}} ligger allerede på enheten med riktig sjekksum',
      bootMenuDesc: 'Last ned netboot.xyz-ISO-en, med kontrollert sjekksum, til den virtuelle CD-en'
    },
    alerts: {
      title: 'Trenger oppmerksomhet',
      temperature: {
        warning: 'Kortet er på {{celsius}} °C. Sjekk at det får luft.',
        critical: 'Kortet er på {{celsius}} °C, som er for varmt. Gi det luft eller slå det av.'
      },
      storage: {
        warning:
          'Bare {{available}} av {{total}} ledig på {{path}}. Store avbildninger får kanskje ikke plass.',
        critical:
          'Bare {{available}} ledig på {{path}}. Opplasting, nedlasting og installasjon av tillegg vil feile. Slett avbildninger du ikke trenger.'
      },
      vpn: '{{name}} skal starte ved oppstart, men kjører ikke, så fjerntilgang gjennom den er nede.',
      openVpn: 'Åpne VPN-innstillinger',
      stream:
        'Videostrømmen har feilet. Prøv en annen videomodus i Skjerm-menyen, eller last inn siden på nytt.'
    },
    power: {
      resetDesc: 'Starter verten på nytt med en gang. Ulagret arbeid går tapt.',
      powerShortDesc: 'Slår på verten, eller ber operativsystemet slå seg av (ACPI).',
      powerLongDesc: 'Tvinger verten av uten avslutning.',
      hddLed: 'Disk-LED',
      hddActive: 'Aktiv',
      hddIdle: 'Inaktiv',
      title: 'På-knapp',
      showConfirm: 'Bekreftelse',
      showConfirmTip:
        'Spør før et kort trykk på av/på. Tilbakestilling og langt trykk spør alltid.',
      reset: 'Reset-knapp',
      power: 'På-knapp',
      powerShort: 'På-knapp (kort trykk)',
      powerLong: 'På-knapp (langt trykk)',
      resetConfirm: 'Fortsette tilbakestilling?',
      powerConfirm: 'Fortsette strømdrift?',
      okBtn: 'Ja',
      cancelBtn: 'Nei',
      hostOs: 'Vertens OS',
      hostOsTip: 'Sendes som USB-taster. Verten bestemmer hva de gjør.',
      sleep: 'Hvilemodus',
      wake: 'Vekk',
      wakeKey: 'Vekk med Shift',
      powerDown: 'Slå av',
      sleepConfirm: 'Sette verten i hvilemodus?',
      powerDownConfirm: 'Sende av-tasten til verten?',
      wakeTip:
        'En vert i hvilemodus ignorerer ofte Vekk fra enheten som satte den i hvile. Vekk med Shift trykker en tast på tastaturet, som flere verter godtar.',
      led: 'Strøm-LED',
      ledOn: 'På',
      ledOff: 'Av',
      ledUnknown: 'Ukjent',
      ledConnected: 'Strøm-LED tilkoblet',
      ledConnectedTip:
        'Slå på bare hvis vertens strøm-LED-kontakt er koblet til kortet. Uten den er strømtilstanden ukjent.',
      ledConnectedFailed: 'Kunne ikke lagre innstillingen for strøm-LED',
      powerLongConfirm:
        'Holde av/på-knappen i {{seconds}} s? Dette kutter strømmen uten avslutning.',
      done: 'Knappen er trykket',
      failed: 'Trykket på knappen mislyktes'
    },
    settings: {
      title: 'Innstillinger',
      nav: {
        system: 'System',
        network: 'Nettverk',
        access: 'Tilgang',
        integrations: 'Integrasjoner',
        boot: 'Oppstart og medier',
        browser: 'Denne nettleseren',
        search: 'Finn en innstilling',
        noMatch: 'Ingen innstillinger passer',
        locked:
          'En operasjon pågår. Andre sider og lukking er ikke tilgjengelig før den er ferdig.',
        vpnProvider: 'VPN-leverandør'
      },
      mcp: {
        keyNote:
          'MCP bruker sin egen API-nøkkel, vist nedenfor. Nøkler fra siden API-nøkler virker ikke her.',
        title: 'MCP-tjeneste',
        service: 'MCP-fjernstyring',
        serviceDesc: 'La klarerte MCP-klienter styre tastatur og mus og ta skjermbilder',
        securityWarning:
          'Alle med denne API-nøkkelen kan styre den eksterne verten og se skjermen. Bruk HTTPS, og aktiver tjenesten bare på klarerte nettverk.',
        endpoint: 'Endepunkt',
        apiKey: 'API-nøkkel',
        regenerateConfirmTitle: 'Generere MCP API-nøkkelen på nytt?',
        regenerateConfirmDesc: 'Den gjeldende nøkkelen slutter å virke umiddelbart.',
        enableConfirmTitle: 'Aktivere ekstern MCP-styring?',
        enableConfirmDesc:
          'Aktivering av MCP stopper PicoClaw og lukker alle aktive PicoClaw-økter.',
        failed: 'MCP-operasjonen mislyktes',
        copyFailed: 'Kopiering mislyktes. Kopier manuelt.',
        okBtn: 'Bekreft',
        cancelBtn: 'Avbryt',
        showKey: 'Vis nøkkel',
        hideKey: 'Skjul nøkkel',
        regenerateKey: 'Lag ny nøkkel'
      },
      redfish: {
        example: 'Eksempel',
        title: 'Redfish',
        service: 'Redfish-tjeneste',
        serviceDesc:
          'DMTF Redfish API, for strømstyring, virtuelle medier og status fra verktøy som redfishtool og Ansible. Når den slås av, avsluttes alle Redfish-økter.',
        endpoint: 'Tjenesterot',
        httpsOn: 'Kortet bruker HTTPS, som de fleste Redfish-verktøy krever.',
        httpsOff:
          'Kortet bruker vanlig HTTP. De fleste Redfish-verktøy krever HTTPS: slå det på i "Innstillinger > Nettverk".',
        credentials:
          'Redfish bruker KVM-kontoene, med Basic-autentisering eller en Redfish-økt, og API-nøkler sendt som X-Auth-Token. API-nøkler administreres på siden API-nøkler.',
        powerActions: 'Strømhandlinger',
        powerActionsDesc:
          'Tilbakestillingstypene som tilbys nå. On, ForceOff og GracefulShutdown trenger strømtilstanden, så de tilbys bare når "Strøm-LED tilkoblet" er slått på i strømmenyen.',
        sessions: 'Økter',
        noSessions: 'Ingen åpne Redfish-økter',
        created: 'Opprettet',
        lastUsed: 'Sist brukt',
        refresh: 'Oppdater',
        end: 'Avslutt',
        endConfirmTitle: 'Avslutte denne Redfish-økten?',
        endConfirmDesc: 'Tokenet slutter å virke umiddelbart. Klienten må logge inn på nytt.',
        failed: 'Redfish-operasjonen mislyktes',
        copyFailed: 'Kopiering mislyktes. Kopier manuelt.',
        okBtn: 'Bekreft',
        cancelBtn: 'Avbryt'
      },
      ipmi: {
        copyBeforeSave: 'Kopier passordet nå. Når det er lagret, kan det ikke vises igjen.',
        noLogin:
          'IPMI er på, men ingen aktiv konto har et IPMI-passord, så ingen kan logge inn. Angi et nedenfor.',
        title: 'IPMI',
        warning:
          'IPMI-autentisering er svak av konstruksjon. Alle som når kortet og kjenner et brukernavn, kan hente en hash av brukerens IPMI-passord og prøve å knekke den offline. Bruk genererte passord, slå på IPMI bare på et nettverk du stoler på, og velg heller Redfish over HTTPS der verktøyet støtter det.',
        service: 'IPMI over LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) på UDP-port 623, for strøm og status på verten. IPMI 1.5 og cipher suite 0 avvises. Når det slås av, avsluttes alle IPMI-økter.',
        example: 'Eksempel',
        copyFailed: 'Kopiering mislyktes. Kopier manuelt.',
        ledOn: 'Strømstatus, on, off, soft, cycle og reset er tilgjengelige.',
        ledOff:
          '"Strøm-LED tilkoblet" er av i strømmenyen, så strømtilstanden er ukjent. Bare "power reset" virker: status, on, off, soft og cycle avvises.',
        accounts: 'Kontoer',
        accountsDesc:
          'IPMI logger inn med KVM-kontoene, hver med sitt eget IPMI-passord, atskilt fra nettpassordet. Administratorer får ADMINISTRATOR. Brukere får USER: de kan lese strømtilstanden med "-L USER", men ikke endre den.',
        passwordSet: 'IPMI-passord satt',
        passwordNotSet: 'Intet IPMI-passord: kan ikke logge inn over IPMI',
        nameTooLong: 'Navnet er lengre enn 16 tegn, noe IPMI ikke tillater',
        accountDisabled: 'Kontoen er deaktivert',
        setPassword: 'Angi passord',
        changePassword: 'Endre passord',
        remove: 'Fjern',
        removeConfirmTitle: 'Fjerne IPMI-passordet til {{user}}?',
        removeConfirmDesc:
          'Kontoen kan ikke lenger logge inn over IPMI, og IPMI-øktene dens avsluttes.',
        passwordTitle: 'IPMI-passord for {{user}}',
        passwordDesc:
          '12 til 20 skrivbare ASCII-tegn, forskjellig fra nettpassordet. IPMI krever at kortet lagrer passordet i en form det kan lese igjen, så bruk et som ikke brukes noe annet sted. Kopier det før du lagrer: det vises ikke igjen.',
        passwordPlaceholder: 'IPMI-passord',
        generate: 'Generer',
        copy: 'Kopier',
        save: 'Lagre',
        passwordLength: 'Bruk 12 til 20 tegn.',
        passwordChars: 'Bruk bare skrivbare ASCII-tegn.',
        saved: 'IPMI-passord lagret',
        failed: 'IPMI-operasjonen mislyktes',
        okBtn: 'Bekreft',
        cancelBtn: 'Avbryt'
      },
      ssh: {
        service: 'SSH-server',
        serviceDesc: 'Start sshd nå og ved hver oppstart',
        failed: 'Kunne ikke laste SSH-innstillingene',
        rootDefault: 'root har fortsatt fabrikkpassordet',
        rootEmpty: 'root har ikke noe passord',
        rootWarning:
          'Alle som når konsollen eller SSH, kan logge inn som root. Angi et passord under {{account}} > {{password}}: for enhetens eier setter det også passordet til root.',
        connection: 'Tilkobling',
        command: 'Logg inn som root',
        port: 'Port',
        viaVpn: 'Via {{name}}',
        notRunning: 'sshd kjører ikke. Slå på SSH-serveren for å koble til.',
        hostKeys: 'Fingeravtrykk for vertsnøkler',
        hostKeysDesc: 'Sammenlign dem med det ssh viser ved første tilkobling.',
        noHostKeys: 'Ingen vertsnøkler ennå. sshd lager dem første gang den starter.',
        keys: 'Autoriserte nøkler',
        keysDesc:
          'Offentlige nøkler som kan logge inn som root. De lagres på datapartisjonen, så oppdateringer beholder dem.',
        noKeys: 'Ingen autoriserte nøkler ennå.',
        noComment: 'ingen kommentar',
        addPlaceholder:
          'Lim inn én offentlig nøkkel, for eksempel innholdet i ~/.ssh/id_ed25519.pub',
        add: 'Legg til nøkkel',
        added: 'Nøkkel lagt til',
        removed: 'Nøkkel fjernet',
        deleteConfirm: 'Fjerne denne nøkkelen?',
        deleteConfirmDesc: 'Den kan ikke lenger logge inn. Åpne økter forblir åpne.',
        invalidKey: 'Dette er ikke en offentlig nøkkel. Lim inn én linje fra en .pub-fil.',
        keyOptions: 'Nøkler med valg som command= eller from= godtas ikke her.',
        duplicateKey: 'Denne nøkkelen er allerede autorisert.',
        lastKey: 'Den siste nøkkelen kan ikke fjernes mens innlogging kun med nøkler er på.',
        keysOnly: 'Kun nøkler',
        keysOnlyDesc:
          'Slå av innlogging med passord og keyboard-interactive. Åpne økter forblir åpne.',
        keysOnlyNeedsKey: 'Legg til en autorisert nøkkel først, ellers kunne ingen logge inn.',
        keysOnlyOn: 'Innlogging med passord slått av',
        keysOnlyOff: 'Innlogging med passord slått på',
        notHonoured:
          'sshd i dette bildet leser ikke denne innstillingen, så innlogging med passord forblir på.',
        reloadFailed:
          'Lagret, men sshd kunne ikke lastes inn på nytt. Det gjelder neste gang sshd starter.',
        notApplied:
          'sshd godtar fortsatt passord. Slå SSH-serveren av og på for å ta i bruk innstillingen.',
        changePort: 'Endre',
        portConfirm: 'Endre SSH-porten til {{port}}?',
        portConfirmDesc:
          'De nåværende SSH-øktene dine forblir åpne. Nye tilkoblinger må bruke port {{port}}. Sørg for at brannmuren tillater det.',
        portChanged: 'SSH-porten er endret til {{port}}',
        portInvalid: 'Oppgi en port fra 1 til 65535.',
        portReserved: 'IronKVM bruker selv denne porten. Velg en annen.',
        portInUse: 'Et annet program på IronKVM lytter allerede på denne porten.',
        portNotHonoured: 'sshd i dette bildet leser ikke denne innstillingen, så porten er uendret.'
      },
      vnc: {
        address: 'Adresse',
        certHint:
          'VeNCrypt X509Plain bruker enhetens selvsignerte sertifikat, så klienten advarer ved første tilkobling. Godta det, eller lagre sertifikatet fra denne sidens HTTPS-adresse og gi det til TigerVNC med -X509CA=<fil>.',
        title: 'VNC',
        service: 'VNC-server',
        serviceDesc:
          'Lar en VNC-klient, for eksempel TigerVNC eller Remmina, se og styre verten. Klienten må støtte Tight-koding. Én økt om gangen.',
        credentials:
          'Logg inn med en KVM-konto. Tilkoblingen krypteres med kortets TLS-sertifikat (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'TCP-porten som serveren lytter på.',
        maxFps: 'Grense for bildefrekvens',
        maxFpsDesc: 'Det høyeste antallet bilder i sekundet som en klient får.',
        vncAuth: 'Enkel VNC-autentisering',
        vncAuthDesc:
          'For klienter uten VeNCrypt. Den sjekker et eget VNC-passord i stedet for en konto.',
        vncAuthWarning:
          'Enkel VNC-autentisering krypterer ikke tilkoblingen. Alle på nettverksveien kan se skjermen og tastetrykkene. Bruk den bare på et klarert nettverk.',
        password: 'VNC-passord',
        passwordSet: 'Et passord er angitt. Skriv inn et nytt for å endre det.',
        passwordInvalid: 'VNC-passordet må være 6 til 8 tegn.',
        save: 'Lagre',
        saved: 'Innstillinger lagret',
        state: 'Status',
        listening: 'Lytter på port {{port}}',
        notListening: 'Lytter ikke',
        noSession: 'Ingen åpen økt',
        client: 'Klient',
        user: 'Bruker',
        method: 'Autentisering',
        methodVencrypt: 'Konto over TLS',
        methodVnc: 'VNC-passord',
        since: 'Tilkoblet siden',
        resolution: 'Oppløsning',
        framesSent: 'Sendte bilder',
        lastError: 'Den siste økten ble avsluttet: {{error}}',
        refresh: 'Oppdater',
        disconnect: 'Koble fra',
        disconnectConfirmTitle: 'Avslutte VNC-økten?',
        disconnectConfirmDesc:
          'Klienten kobles fra med en gang, og alle taster og knapper den holder nede, slippes.',
        failed: 'VNC-operasjonen mislyktes',
        okBtn: 'Bekreft',
        cancelBtn: 'Avbryt'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Vert-watchdog',
        serviceDesc:
          'Hvis verten skal være på, og bildet ikke endrer seg, eller det ikke er noe HDMI-signal, i hele tidsavbruddet, trykker kortet på reset eller slår verten av og på.',
        stillWarning:
          'En vert der skjermen går i hvilemodus, eller der bildet står stille mens den arbeider, ser ut til å ha hengt seg. Slå av skjermdvale på verten, eller angi en ping-adresse.',
        ledHint:
          '"Strøm-LED tilkoblet" er slått av i strømmenyen. Watchdogen ser ikke når verten er av, så den behandler verten som alltid på.',
        timeout: 'Tidsavbrudd',
        timeoutDesc: 'Hvor lenge verten kan være uten livstegn før watchdogen griper inn.',
        action: 'Handling',
        actionDesc: 'Av og på holder av/på-knappen inne i 5 sekunder og trykker den så inn igjen.',
        actionReset: 'Reset-knapp',
        actionPower: 'Av og på',
        cooldown: 'Pause',
        cooldownDesc: 'Den korteste tiden mellom to handlinger.',
        maxPerHour: 'Handlinger per time',
        maxPerHourDesc: 'Det høyeste antallet handlinger i løpet av en time.',
        pingHost: 'Ping-adresse',
        pingHostDesc:
          'IP-adressen til verten. Et svar teller som et livstegn. La feltet stå tomt for å ikke pinge.',
        pingHostInvalid: 'Skriv inn en IPv4- eller IPv6-adresse.',
        minutes: 'min',
        save: 'Lagre',
        saved: 'Lagret',
        state: 'Detektor',
        status: {
          off: 'Av',
          watching: 'Overvåker',
          hostOff: 'Vert av',
          captureOff: 'HDMI-opptak av',
          cooldown: 'Pause',
          capped: 'Timegrense nådd',
          acting: 'Griper inn'
        },
        signal: 'HDMI-signal',
        yes: 'Ja',
        no: 'Nei',
        led: 'Strøm-LED',
        on: 'På',
        off: 'Av',
        ledNotConnected: 'Ikke tilkoblet',
        ping: 'Ping',
        pingNotSet: 'Ikke angitt',
        pingReply: 'Svarer',
        pingNoReply: 'Ingen svar',
        lastChange: 'Siste bildeendring',
        never: 'Aldri',
        actsIn: 'Griper inn om',
        actionsLastHour: 'Handlinger siste time',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Logg',
        noLog: 'Watchdogen har ikke grepet inn ennå.',
        refresh: 'Oppdater',
        reasonFrozen: 'Bildet endret seg ikke',
        reasonNoSignal: 'Ingen HDMI-signal',
        stuckFor: 'ingen livstegn på {{duration}}',
        pressFailed: 'Trykket mislyktes: {{error}}',
        noScreenshot: 'Ingen skjermbilde',
        failed: 'Watchdog-handlingen mislyktes',
        powerNeedsLed: 'Strømsyklus krever «Strøm-LED tilkoblet» i strømmenyen.',
        noLedConfirmTitle: 'Slå på watchdog uten strøm-LED?',
        noLedConfirmDesc:
          'Kortet kan ikke se når verten er av, så det behandler verten som alltid på. Hvis du slår av verten, trykker watchdog på tilbakestill når tidsavbruddet har gått. Koble til strøm-LED-en for å unngå dette.',
        noLedConfirmOk: 'Slå på',
        cancel: 'Avbryt'
      },
      media: {
        title: 'Virtuelle medier',
        description:
          'Oppsett for dialogen Medier i verktøylinjen. Montering av avbildninger, å legge dem til og valg av Ventoy-settet skjer i dialogen.',
        ejectFirst:
          'Ventoy-disken står i en stasjon. Løs den ut i dialogen Medier for å avinstallere.'
      },
      netboot: {
        title: 'Nettverksoppstart',
        isoDownload: 'Last ned',
        description:
          'Start verten fra nettverket: iPXE og en meny med avbildningene på KVM-en over USB-nettverkstilkoblingen, eller netboot.xyz via proxy-DHCP på LAN-et.',
        addon: 'dnsmasq og oppstartsfiler',
        addonDesc:
          'Installert på /data: dnsmasq fra Alpine, iPXE og netboot.xyz fra utgivelsene deres, hver kontrollert mot sin sjekksum.',
        install: 'Installer',
        installing: 'Installerer. Dette kan ta noen minutter.',
        uninstall: 'Avinstaller',
        uninstallConfirm: 'Slå av nettverksoppstart og fjerne dnsmasq og oppstartsfilene?',
        needsData: 'Nettverksoppstart krever et IronKVM-bilde med /data-partisjonen montert.',
        usb: 'På USB-nettverkstilkoblingen',
        usbDesc:
          'Mens USB-nettverkstilkoblingen er på, betjener dnsmasq den i stedet for udhcpd. Verten får sin ene adresse uten ruter og uten DNS-server, iPXE for sin arkitektur og en meny med ISO-avbildningene på KVM-en.',
        linkOff: 'USB-nettverkstilkoblingen er av. Slå den på under Enhet, USB-nettverk.',
        menuUrl: 'Meny',
        leases: 'Vertens lease',
        noLeases: 'Ingen ennå',
        netbootxyzNote:
          'netboot.xyz i menyen lastes fra internett, som USB-tilkoblingen ikke når. Verten trenger internett på en annen nettverksport.',
        lan: 'Proxy-DHCP på LAN-et',
        lanDesc:
          'Svarer PXE-klienter på LAN-et med netboot.xyz, som deretter laster menyen sin fra internett. Den deler aldri ut adresser og tilbyr ikke avbildningene på KVM-en.',
        lanWarning:
          'Alle PXE-klienter på dette LAN-et får tilbud om netboot.xyz, ikke bare verten. Slå dette på bare i et nettverk du kontrollerer.',
        lanConfirm: 'Slå på proxy-DHCP på LAN-et?',
        lanInterface: 'LAN',
        running: 'Kjører',
        stopped: 'Kjører ikke',
        images: 'Avbildninger i menyen',
        noImages: 'Ingen ISO-avbildninger i bildemappen.',
        boots: 'Nylige oppstarter',
        noBoots: 'Verten har ikke hentet noe ennå.',
        log: 'dnsmasq-logg',
        refresh: 'Oppdater',
        okBtn: 'Bekreft',
        cancelBtn: 'Avbryt',
        failed: 'Nettverksoppstart-operasjonen mislyktes'
      },
      about: {
        title: 'Om IronKVM',
        information: 'Informasjon',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Applikasjonsversjon',
        applicationTip: 'Versjon av IronKVM-webapplikasjonen',
        image: 'Arkivfil-versjon',
        imageTip: 'IronKVM-kortbilde og NanoKVM-systembildet det bygger på',
        kernel: 'Kjerneversjon',
        kernelTip: 'Versjonen av Linux-kjernen som kjører nå',
        deviceKey: 'Enhetsnøkkel',
        community: 'Fellesskap',
        hostname: 'Vertsnavn',
        hostnameUpdated: 'Vertsnavn oppdatert. Start på nytt for å søke.',
        ipType: {
          Wired: 'Kablet',
          Wireless: 'Trådløs',
          Other: 'Annet'
        },
        hostnameInvalid:
          'Bruk bokstaver, sifre og bindestreker, opptil 63 per punktumdelt del. Ingen bindestrek i starten eller slutten av en del.',
        hostnameFailed: 'Kunne ikke endre vertsnavnet',
        editHostname: 'Rediger vertsnavn',
        docs: 'Dokumentasjon',
        hardware: 'Maskinvare',
        hardwareFaq: 'Maskinvare-FAQ',
        disclaimer:
          'IronKVM: herdet fellesskapsfastvare for Sipeed NanoKVM. Ikke tilknyttet Sipeed.',
        basedOn: 'basert på NanoKVM {{version}}'
      },
      preferences: {
        title: 'Preferanser'
      },
      performance: {
        title: 'Ytelse',
        memory: {
          title: 'Minne',
          description: 'RAM, swap og hva som bruker dem. Oppdateres med noen sekunders mellomrom.',
          ram: 'RAM',
          of: '{{used}} av {{total}}',
          available: '{{available}} ledig',
          availableLow: 'Bare {{available}} ledig. Tjenester kan bli trege eller bli stoppet.',
          swap: 'Swap',
          swapFile: 'Swapfil',
          zram: 'Komprimert swap',
          zramRam: '{{ram}} i RAM',
          consumers: 'Største forbrukere',
          addons: 'Tillegg',
          addonsTip: 'Minnegruppen Tailscale og NetBird kjører i, målt mot grensen.',
          video: 'Videominne',
          videoTip: 'Minne reservert for videoopptak. Det deles ikke med resten av systemet.',
          videoGenerations_one: '{{count}} tidligere IronKVM-økt holder på videominne',
          videoGenerations_other: '{{count}} tidligere IronKVM-økter holder på videominne',
          videoReboot: 'Start på nytt for å frigjøre det.'
        }
      },
      appearance: {
        thisBrowser: 'Denne nettleseren',
        thisBrowserDesc: 'Lagres bare i denne nettleseren. Andre nettlesere har sine egne.',
        deviceWide: 'Enhet',
        deviceWideDesc: 'Lagres på enheten. Gjelder for alle som åpner den.',
        language: 'Språk',
        languageDesc: 'Velg språket for grensesnittet',
        webTitle: 'Netttittel',
        webTitleDesc: 'Tilpass nettsidetittelen',
        menuBar: {
          title: 'Menylinje',
          mode: 'Visningsmodus',
          modeDesc: 'Vis menylinje på skjermen',
          modeOff: 'Av',
          modeAuto: 'Skjul automatisk',
          modeAlways: 'Alltid synlig',
          keyboardLedStatus: 'Indikatorer for tastaturlås',
          keyboardLedStatusDesc:
            'Vis Num Lock-, Caps Lock- og Scroll Lock-status for den eksterne datamaskinen',
          icons: 'Undermenyikoner',
          iconsDesc: 'Vis undermenyikoner i menylinjen'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Status for låser på eksternt tastatur',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'På',
        off: 'Av',
        unknown: 'Ukjent'
      },
      device: {
        title: 'Enhet',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'OLED-lysstyrke',
          brightnessDescription: 'Et lavere nivå gir skjermen lengre levetid',
          brightnessLevels: {
            '64': 'Lavest',
            '96': 'Lav',
            '128': 'Middels',
            '160': 'Høy',
            '207': 'Standard',
            '255': 'Maksimal'
          },
          0: 'Aldri',
          15: '15 sek',
          30: '30 sek',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 time'
        },
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Frontpanel'
        },
        cpuFreq: {
          title: 'CPU-frekvens',
          description: 'Angi CPU-klokken som brukes ved neste oppstart',
          tip: 'CPU-en starter på 850 MHz og er spesifisert for 1000 MHz. En ny verdi tas i bruk ved neste oppstart, ikke mens systemet kjører. 1000 MHz er innenfor spesifikasjonen; temperaturen er godt innenfor grensene med begge innstillingene.',
          running: 'Kjører: {{mhz}} MHz',
          rebootToApply: 'start på nytt for å ta i bruk',
          rebootConfirm: 'Starte på nytt nå for å ta i bruk {{mhz}} MHz?'
        },
        swap: {
          title: 'Bytt',
          disable: 'Deaktiver',
          description: 'Angi størrelsen på byttefilen',
          tip: 'Aktivering av denne funksjonen kan forkorte SD-kortets brukbare levetid!',
          active: 'Aktiv - {{used}} av {{total}}',
          inactive: 'Satt opp, men ikke i bruk'
        },
        zram: {
          title: 'Komprimert swap (zram)',
          description: 'Swap i komprimert RAM i stedet for på SD-kortet',
          tip: 'zram holder swap unna SD-kortet, så det gir ingen slitasje. Det finnes ingen disk-swap bak: hvis zram blir full, stopper kjernen en prosess i stedet for å swappe sakte. Minnegrensen begrenser hvor mye RAM zram kan bruke.',
          unavailable: 'Kjernemodulene er ikke installert på denne enheten',
          inactive: 'Aktivert, men enheten startet ikke',
          active: 'Aktiv - {{used}} av {{total}}, {{ratio}}x',
          off: 'Av',
          detail: {
            algorithm: 'Algoritme: {{algorithm}}',
            memory: 'Minne brukt: {{used}} av {{limit}}',
            memoryNoLimit: 'Minne brukt: {{used}}, ingen grense satt',
            counters: 'Sider swappet inn {{in}}, ut {{out}} (alle swap-enheter, siden oppstart)'
          }
        },
        mouseJiggler: {
          title: 'Mus Jiggler',
          description: 'Hindre den eksterne verten fra å sove',
          disable: 'Deaktiver',
          absolute: 'Absolutt modus',
          relative: 'Relativ modus'
        },
        mdns: {
          description: 'Aktiver mDNS oppdagelsestjeneste',
          tip: 'Slå den av hvis den ikke er nødvendig'
        },
        hdmi: {
          description: 'Aktiver HDMI/skjermutgang',
          idleTimeoutTitle: 'Tidsavbrudd for inaktivt opptak',
          idleTimeoutDescription: 'Stopp HDMI-opptak etter at det ikke har vært aktive seere i',
          minutes: 'min'
        },
        hidOnly: 'HID-Bare modus',
        hidOnlyDesc: 'Slutt å emulere virtuelle enheter, behold bare grunnleggende HID-kontroll',
        disk: 'Virtuell disk',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Virtuelt nettverk',
        networkDesc: 'Monter virtuelt nettverkskort på den eksterne verten',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Vert:',
          description:
            'En privat nettverkstilkobling til den eksterne verten via USB-kabelen. Verten får en adresse uten gateway og uten DNS, så den kan ikke nå ditt lokalnett gjennom IronKVM.',
          mode: 'Protokoll',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (for verter uten NCM)',
          rndis: 'RNDIS (tilbys ikke lenger)',
          rndisNote: 'Denne tilkoblingen bruker RNDIS, som ikke tilbys lenger. Velg NCM eller ECM.',
          subnet: 'Delnett',
          subnetDesc:
            'Et privat IPv4-nettverk, /24 til /30. IronKVM tar den første adressen, verten den andre.',
          invalidSubnet: 'Skriv inn et delnett, for eksempel 172.31.255.0/30.',
          apply: 'Bruk',
          confirm: 'Koble til USB-enheten på nytt?',
          reenumerate:
            'Når du bruker endringen, bygges USB-tilkoblingen opp på nytt. Verten mister tastatur, mus og virtuell disk i noen sekunder.'
        },
        audio: 'Virtuell høyttaler',
        audioDesc:
          'Vis et USB-lydkort for den eksterne verten, slik at du kan høre den. Verten må velge det som utenhet for lyd. Endring av dette bygger opp USB-tilkoblingen på nytt.',
        audioNote: 'Lyd er tilgjengelig i begge H.264-modusene (WebRTC og Direct), ikke i MJPEG',
        console: 'Seriekonsoll',
        consoleDesc:
          'Vis en USB-serieport for den eksterne verten, for å logge inn på denne IronKVM når nettverket ikke er tilgjengelig',
        consoleTip:
          'Alle som kontrollerer den eksterne verten får en innloggingsforespørsel til denne IronKVM. Angi et sterkt passord før du aktiverer (Konto - Endre passord).',
        usbApply: {
          changed: 'Endret',
          discard: 'Forkast',
          pending: 'Endringene er ikke tatt i bruk ennå.'
        },
        endpoints: {
          title: 'USB-plasser',
          free: '{{free}} av {{total}} ledige',
          slots: 'Plasser: {{count}}',
          full: 'Ikke nok ledige USB-plasser. Slå av noe annet først.',
          inactive:
            'På, men kjører ikke: USB-kontrolleren gikk tom for plasser. Slå av en annen enhet, så starter denne med en gang.',
          explain:
            'USB-kontrolleren har et fast antall plasser (inngående endepunkter), og tastatur og mus tar alltid noen. Er flere enheter på enn det er plass til, beholdes tastatur og mus, og resten slås av.',
          error: 'Fikk ikke kontakt med enheten. Prøv igjen.',
          fitTogether: 'Disse passer sammen: {{sets}}'
        },
        reboot: 'Start på nytt',
        rebootDesc: 'Er du sikker på at du vil starte IronKVM på nytt?',
        okBtn: 'Ja',
        cancelBtn: 'Nei',
        rebootFailed: 'Omstart mislyktes'
      },
      network: {
        title: 'Nettverk',
        wifi: {
          disconnectBtn: 'Koble fra',
          disconnectWarning:
            'Hvis du når IronKVM via dette Wi-Fi-nettverket, mister denne siden tilkoblingen.',
          disconnected: 'Wi-Fi frakoblet',
          title: 'Wi-Fi',
          description: 'Konfigurer Wi-Fi',
          apMode: 'AP-modus er aktivert, koble til Wi-Fi ved å skanne QR-koden',
          connect: 'Koble til Wi-Fi',
          connectDesc1: 'Skriv inn nettverkets SSID og passord',
          connectDesc2: 'Skriv inn passordet for å koble til dette nettverket',
          disconnect: 'Er du sikker på at du vil koble fra nettverket?',
          failed: 'Tilkobling mislyktes, prøv igjen.',
          ssid: 'Navn',
          password: 'Passord',
          joinBtn: 'Koble til',
          confirmBtn: 'OK',
          cancelBtn: 'Avbryt'
        },
        tls: {
          description: 'Aktiver HTTPS-protokoll',
          tip: 'Merk: Bruk av HTTPS kan øke forsinkelsen, spesielt i MJPEG-videomodus.',
          restarting: 'Starter enhetsserveren på nytt, dette tar omtrent to minutter...',
          waiting: 'Venter på at enheten svarer igjen...',
          waitingHttp:
            'Bytter tilbake til http. Last inn siden på nytt hvis den ikke åpnes av seg selv.',
          failed: 'Kunne ikke endre HTTPS-innstillingen',
          enableConfirm: 'Slå på HTTPS?',
          disableConfirm: 'Slå av HTTPS?',
          confirmDesc:
            'Dette logger deg ut og starter enhetens server på nytt, noe som tar omtrent to minutter. Siden åpner deretter {{url}}.',
          confirmOk: 'Fortsett',
          confirmCancel: 'Avbryt'
        },
        ethernet: {
          title: 'IP-adresse',
          description: 'Konfigurer hvordan IronKVM får adressen sin på det kablede nettverket',
          dhcp: 'DHCP',
          manual: 'Manuell',
          networkDetails: 'Nettverksdetaljer',
          interface: 'Grensesnitt',
          ipAddress: 'IP-adresse',
          subnetMask: 'Nettverksmaske',
          router: 'Ruter',
          save: 'Bruk',
          invalidAddress: 'Skriv inn en gyldig IP-adresse',
          invalidMask: 'Skriv inn en gyldig nettverksmaske, for eksempel 255.255.255.0 eller 24',
          invalidRouter: 'Skriv inn en gyldig ruteradresse',
          addressRequired: 'En IP-adresse er påkrevd',
          maskRequired: 'En nettverksmaske er påkrevd',
          applyTitle: 'Vil du endre adressen til IronKVM?',
          applyWarning:
            'Forbindelsen til denne siden går tapt. IronKVM tar i bruk den nye adressen og venter {{seconds}} sekunder på at du når den der. Å nå den beholder endringen. Hvis ingenting når den, gjenoppretter IronKVM de forrige innstillingene.',
          applyConfirm: 'Bruk',
          applyCancel: 'Avbryt',
          applyFailed: 'Adressen kunne ikke tas i bruk',
          trialTitle: 'Venter på bekreftelse',
          trialDhcp: 'IronKVM ber om en adresse via DHCP.',
          trialStatic: 'IronKVM er nå på {{address}}.',
          trialInstruction:
            'Åpne IronKVM på den nye adressen, og logg inn hvis den ber om det. Å nå den der beholder endringen. Hvis ingenting når IronKVM innen {{seconds}} sekunder, gjenoppretter den de forrige innstillingene.',
          trialOpen: 'Åpne den nye adressen',
          trialKeep: 'Behold disse innstillingene',
          trialKept: 'Den nye adressen er lagret',
          trialKeepFailed: 'Innstillingene kunne ikke beholdes',
          trialGone: 'Endringen er allerede tilbakestilt. Prøv igjen.',
          unsaved: 'Ulagrede endringer'
        },
        dns: {
          title: 'DNS',
          description: 'Konfigurer DNS-servere for IronKVM',
          mode: 'Modus',
          dhcp: 'DHCP',
          manual: 'Manuell',
          add: 'Legg til DNS',
          save: 'Lagre',
          invalid: 'Skriv inn en gyldig IP-adresse',
          noDhcp: 'Ingen DHCP-DNS er tilgjengelig nå',
          saved: 'DNS-innstillinger lagret',
          saveFailed: 'Kunne ikke lagre DNS-innstillinger',
          unsaved: 'Ulagrede endringer',
          maxServers: 'Maksimalt {{count}} DNS-servere er tillatt',
          dnsServers: 'DNS-servere',
          dhcpServersDescription: 'DNS-servere hentes automatisk fra DHCP',
          manualServersDescription: 'DNS-servere kan redigeres manuelt',
          networkDetails: 'Nettverksdetaljer',
          interface: 'Grensesnitt',
          ipAddress: 'IP-adresse',
          subnetMask: 'Subnettmaske',
          router: 'Ruter',
          none: 'Ingen'
        },
        syslog: {
          title: 'Fjernlogging',
          description: 'Send system-, kjerne- og IronKVM-logger til en syslog-mottaker over UDP.',
          placeholder: 'vert eller vert:port, UDP, standardport 514',
          save: 'Lagre',
          turnOff: 'Slå av',
          test: 'Send testmelding',
          sent: 'Sendt <v>{{message}}</v>. Søk etter den i mottakeren.',
          forwarding: 'Videresender til <v>{{target}}</v>',
          local: 'Bare lokalt',
          savedButLocal: 'Lagret <v>{{target}}</v>, men den kjørende loggeren logger bare lokalt',
          savedButForwarding:
            'Lagret <v>{{target}}</v>, men den kjørende loggeren videresender til <v>{{active}}</v>',
          offButForwarding:
            'Slått av, men den kjørende loggeren videresender fortsatt til <v>{{active}}</v>',
          unsupported: 'Loggeren i dette bildet ignorerer innstillingen; oppdater for å bruke den',
          loadFailed: 'Kunne ikke lese logginnstillingen',
          saveFailed: 'Kunne ikke lagre logginnstillingen',
          testFailed: 'Kunne ikke sende testmeldingen',
          metrics: 'Metrikker',
          metricsDesc:
            'Prometheus kan hente URL-en nedenfor med et bearer-token fra <link>API-nøkler</link>.',
          metricsUrl: 'Metrikk-URL',
          errors: {
            empty: 'Skriv inn en vert eller vert:port',
            long: 'For langt for et vertsnavn',
            ipv6: 'Sett en IPv6-adresse i hakeparenteser, som [fd00::1]:514',
            brackets: 'Adressen i parentesene er ikke en IPv6-adresse',
            host:
              'Bruk bare bokstaver, sifre, punktum og bindestreker, og ikke start med punktum eller bindestrek',
            port: 'Porten må være et tall fra 1 til 65535'
          }
        }
      },
      vpn: {
        kvmUrl: 'KVM-adresse',
        moreTip: 'Flere handlinger',
        updateTip: 'Oppdater til {{version}}',
        loading: 'Laster...',
        okBtn: 'Ja',
        cancelBtn: 'Nei',
        restart: 'Starte {{name}} på nytt?',
        update: 'Oppdatere {{name}} til {{version}}?',
        updateDesc: 'Tjenesten starter på nytt hvis den kjører. Innloggingen beholdes.',
        notInstall: '{{name}} er ikke installert.',
        install: 'Installer',
        installing: 'Installerer',
        installFailed: 'Installasjonen mislyktes',
        retry: 'Prøv igjen',
        connected: 'Tilkoblet',
        connectedDesc:
          'På kobler til {{name}}-nettverket. Av kobler fra og stopper {{name}} for å frigjøre minne.',
        connectAtBoot: 'Koble til ved oppstart',
        connectAtBootDesc: 'Koble til {{name}}-nettverket når KVM starter.',
        restartService: 'Start tjenesten på nytt',
        needsLogin: 'Krever innlogging',
        error: 'Feil',
        thisDevice: 'Denne enheten',
        memoryOf: '{{used}} (grense for tillegg {{limit}})',
        memoryPressed: 'Minnegruppen for tillegg er nær grensen. Det kan hjelpe å slå på swap.',
        peersSummary: 'Noder: {{online}} tilkoblet av {{total}}',
        peersSummaryIdle: 'Noder: {{online}} tilkoblet, {{idle}} ved behov, av {{total}}',
        showOffline: 'Vis frakoblede ({{offline}})',
        hideOffline: 'Skjul frakoblede',
        blocked:
          '{{other}} er på eller kobler til ved oppstart. Bare én VPN kan kjøre om gangen: slå først av Tilkoblet og Koble til ved oppstart for {{other}}.',
        deviceName: 'Enhetsnavn',
        deviceIP: 'Enhetens IP',
        account: 'Konto',
        version: 'Versjon',
        uptime: 'Oppetid',
        noPeers: 'Ingen noder ennå.',
        online: 'Tilkoblet',
        offline: 'Frakoblet',
        idle: 'Ved behov: NetBird kobler til når trafikken trenger det',
        memory: 'Minne',
        uninstall: 'Avinstaller {{name}}',
        uninstallDesc:
          'Er du sikker på at du vil avinstallere {{name}}? Innloggingen blir liggende på kortet.',
        copy: 'Kopier',
        copied: 'Lenke kopiert',
        copyFailed: 'Kunne ikke kopiere lenken. Merk den og kopier den manuelt.',
        open: 'Åpne',
        checkAgain: 'Sjekk igjen',
        notSignedIn: 'Ikke logget inn ennå. Fullfør innloggingen via lenken, og sjekk igjen.',
        checkFailed: 'Kunne ikke sjekke innloggingsstatus',
        loginWaiting:
          'Siden sjekker med noen sekunders mellomrom og fortsetter når du har logget inn.',
        uninstallFailed: 'Avinstallering mislyktes',
        loginFailed: 'Innlogging mislyktes'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Last ned',
        package: 'installasjonspakken',
        unzip: 'og pakk den ut',
        notLogin:
          'Denne enheten er ikke knyttet til din konto enda. Vennligst logg inn og knytt den til kontoen din..',
        urlPeriod: 'Denne lenken er gyldig i 10 minutter',
        login: 'Logg inn',
        logout: 'Logg ut',
        logoutDesc: 'Er du sikker på at du vil logge ut?',
        manualIntro: 'Eller installer manuelt over SSH:',
        copyBinaries: 'Kopier tailscale og tailscaled til {{dir}} på IronKVM',
        linksFile: 'Lag en fil som heter links i samme mappe, med disse to linjene:',
        rebootRefresh: 'Start IronKVM på nytt, og oppdater deretter denne siden'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Denne enheten har ikke blitt med i et NetBird-nettverk ennå. Bli med med en oppsettsnøkkel, eller logg inn med SSO.',
        setupKey: 'Oppsettsnøkkel',
        setupKeyPlaceholder: 'Lim inn en oppsettsnøkkel fra NetBird-dashbordet',
        join: 'Bli med',
        or: 'eller',
        sso: 'Logg inn med SSO',
        urlPeriod: 'Denne lenken er gyldig i 10 minutter',
        logout: 'Avregistrer',
        logoutDesc:
          'Avregistrering fjerner denne noden fra NetBird-kontoen din og sletter konfigurasjonen her. For å bli med igjen trengs en oppsettsnøkkel eller SSO-innlogging, og noden kan få en ny IP. Fortsette?',
        joinFailed: 'Kunne ikke bli med i nettverket'
      },
      update: {
        title: 'Se etter oppdatering',
        queryFailed: 'Kunne ikke hente versjon',
        updateFailed: 'En feil oppstod under oppdatering. Vennligst forsøk igjen.',
        isLatest: 'Du har siste versjon allerede.',
        available: 'En oppdatering er tilgjengelig. Er du sikker på at du ønsker å oppdatere?',
        updating: 'Oppdatering har startet. Vennligst vent...',
        confirm: 'Oppdater',
        cancel: 'Avbryt',
        preview: 'Forhåndsvisningsoppdateringer',
        previewDesc: 'Få tidlig tilgang til nye funksjoner og forbedringer',
        previewTip:
          'Vær oppmerksom på at forhåndsvisningsutgivelser kan inneholde feil eller ufullstendig funksjonalitet!',
        customServer: {
          title: 'Egendefinert oppdateringsserver',
          desc: 'Se etter og last ned nettbaserte oppdateringer fra en angitt server',
          invalidUrl:
            'Angi en gyldig HTTP- eller HTTPS-servermappe uten spørring, fragment eller latest.json.',
          loadFailed: 'Kunne ikke laste inn konfigurasjonen for oppdateringsserveren.',
          saveFailed: 'Kunne ikke lagre konfigurasjonen for oppdateringsserveren.',
          saved: 'Konfigurasjonen for oppdateringsserveren er lagret.',
          save: 'Lagre',
          confirmTitle: 'Vil du bruke en egendefinert oppdateringsserver?',
          confirmDesc:
            'SHA-512 kontrollerer bare at pakken samsvarer med manifestet fra denne serveren. Det beviser ikke at pakken er en offisiell IronKVM-utgivelse. En feilkonfigurert eller ondsinnet server kan gjøre enheten ubrukelig, føre til tap av data eller kompromittere systemet.',
          confirm: 'Bruk likevel',
          useSipeed: 'Bruk den offisielle Sipeed-serveren',
          previewDisabled:
            'Forhåndsvisningsoppdateringer er ikke tilgjengelige mens en egendefinert oppdateringsserver er aktivert.'
        },
        offline: {
          chooseFile: 'Velg fil',
          installing: 'Opplasting fullført. Installerer...',
          noFile: 'Ingen fil valgt',
          title: 'Offline oppdateringer',
          desc: 'Oppdater gjennom lokal installasjonspakke',
          upload: 'Last opp',
          checksumPlaceholder: 'SHA-256-sjekksum (valgfritt)',
          invalidChecksum: 'SHA-256-sjekksummen må inneholde 64 heksadesimale tegn.',
          checksumMismatch: 'SHA-256-verifiseringen mislyktes. Pakken kan være skadet.',
          invalidName: 'Ugyldig filnavnformat. Last ned fra GitHub-utgivelser.',
          updateFailed: 'En feil oppstod under oppdatering. Vennligst forsøk igjen.'
        },
        updateTo: 'Oppdater til {{version}}',
        updateConfirmDesc:
          'Enheten installerer oppdateringen og starter serveren på nytt. Siden lastes inn på nytt når serveren er tilbake.',
        releaseNotes: 'Utgivelsesnotater'
      },
      account: {
        title: 'Konto',
        webAccount: 'Navn på webkonto',
        role: 'Rolle',
        roles: { admin: 'Administrator', user: 'Bruker' },
        password: 'Passord',
        updateBtn: 'Update',
        logoutBtn: 'Logg ut',
        logoutDesc: 'Er du sikker på at du vil logge ut?',
        okBtn: 'Ja',
        cancelBtn: 'Nei',
        users: {
          title: 'Brukere',
          create: 'Opprett bruker',
          enabled: 'Aktivert',
          disabled: 'Deaktivert',
          deviceOwner: 'Enhetseier',
          resetPassword: 'Tilbakestill passord',
          delete: 'Slett',
          deleteConfirm: 'Slette denne brukeren og tilbakekalle alle øktene deres?',
          created: 'Bruker opprettet',
          deleted: 'Bruker slettet',
          passwordUpdated: 'Passord oppdatert',
          loadFailed: 'Kunne ikke laste brukere',
          saveFailed: 'Kunne ikke lagre bruker',
          deleteFailed: 'Kunne ikke slette bruker'
        }
      },
      apiKeys: {
        mcpNote: 'Disse nøklene virker ikke for MCP, som har sin egen nøkkel på MCP-siden.',
        metricsUrl: 'Metrikk-URL',
        monitoring: 'Overvåking',
        monitoringDesc:
          'Prometheus leser metrikkene med en API-nøkkel fra denne siden, sendt som Bearer-token. Alle roller kan lese dem.',
        scrapeConfig: 'Prometheus scrape-konfigurasjon',
        title: 'API-nøkler',
        description:
          'En nøkkel opptrer som eieren sin, med den brukerens rolle. Send den som Authorization: Bearer <key> for målinger og API-et, eller som X-Auth-Token for Redfish.',
        name: 'Navn',
        namePlaceholder: 'Hva nøkkelen er til, for eksempel prometheus',
        nameRequired: 'Gi nøkkelen et navn',
        nameTooLong: 'Navnet kan være maks 64 tegn',
        unnamed: '(uten navn)',
        create: 'Opprett nøkkel',
        created: 'Opprettet',
        owner: 'Eier',
        empty: 'Ingen API-nøkler',
        newKeyTitle: 'Din nye API-nøkkel',
        newKeyWarning:
          'Kopier nøkkelen nå. Den lagres ikke og kan ikke vises igjen. Hvis du mister den, tilbakekall den og opprett en ny.',
        copy: 'Kopier',
        copied: 'Kopiert',
        copyFailed: 'Kopiering mislyktes. Kopier manuelt.',
        done: 'Ferdig',
        revoke: 'Tilbakekall',
        revokeConfirmTitle: 'Tilbakekalle denne API-nøkkelen?',
        revokeConfirmDesc: 'Alt som bruker "{{name}}" slutter å virke umiddelbart.',
        revoked: 'API-nøkkel tilbakekalt',
        loadFailed: 'Kunne ikke laste API-nøkler',
        createFailed: 'Kunne ikke opprette API-nøkkel',
        revokeFailed: 'Kunne ikke tilbakekalle API-nøkkel',
        cancelBtn: 'Avbryt'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistent',
      empty: 'Åpne panelet og start en oppgave for å begynne.',
      inputPlaceholder: 'Beskriv hva du vil at PicoClaw skal gjøre',
      newConversation: 'Ny samtale',
      processing: 'Behandler...',
      agent: {
        defaultTitle: 'Generell assistent',
        defaultDescription: 'Generell chat-, søk- og arbeidsområdehjelp.',
        kvmTitle: 'Fjernstyring',
        kvmDescription: 'Betjen den eksterne verten gjennom IronKVM.',
        switched: 'Agentrolle byttet',
        switchFailed: 'Kunne ikke bytte agentrolle'
      },
      send: 'Send',
      cancel: 'Avbryt',
      status: {
        connecting: 'Kobler til gateway...',
        connected: 'PicoClaw-økt tilkoblet',
        disconnected: 'PicoClaw-økt frakoblet',
        stopped: 'Stoppforespørsel sendt',
        runtimeStarted: 'PicoClaw runtime startet',
        runtimeStartFailed: 'Kunne ikke starte PicoClaw runtime',
        runtimeStopped: 'PicoClaw runtime stoppet',
        runtimeStopFailed: 'Kunne ikke stoppe PicoClaw runtime',
        controlSwitchedToMCP: 'Styringen er byttet til den eksterne MCP-tjenesten'
      },
      connection: {
        runtime: {
          checking: 'Kontrollerer',
          restoring: 'Gjenoppretter PicoClaw',
          ready: 'Runtime klar',
          stopped: 'Runtime stoppet',
          blockedByMCP: 'Ekstern MCP-styring er aktiv',
          readyBlockedByMCP: 'Runtime kjører, men ekstern MCP styrer nå enhetens inndata.',
          readyWithoutControl:
            'Runtime kjører. Gi PicoClaw enhetsstyring før du kobler til på nytt.',
          unavailable: 'Runtime utilgjengelig',
          configError: 'Konfigurasjonsfeil'
        },
        transport: {
          connecting: 'Kobler til',
          connected: 'Tilkoblet',
          disconnected: 'Frakoblet',
          reconnect: 'Koble til på nytt',
          reconnectDescription: 'Koble til den kjørende PicoClaw-økten på nytt.',
          reconnectBlocked: 'PicoClaw trenger enhetsstyring før den kan koble til på nytt.'
        },
        run: {
          idle: 'Inaktiv',
          busy: 'Opptatt'
        }
      },
      message: {
        toolAction: 'Handling',
        observation: 'Observasjon',
        screenshot: 'Skjermbilde'
      },
      overlay: {
        locked: 'PicoClaw kontrollerer enheten. Manuell inntasting er satt på pause.'
      },
      control: {
        picoclaw: 'Enhetsstyring: PicoClaw',
        picoclawDescription:
          'PicoClaw kan sende tastatur- og musinndata. Manuell inndata kan bli satt på pause.',
        mcp: 'Enhetsstyring: ekstern MCP',
        mcpDescription: 'Ekstern MCP kan skrive til enheten. PicoClaw tar ikke over inndata.',
        off: 'Enhetsstyring: av',
        offDescription:
          'KI sender ikke tastatur- eller musinndata. Manuell styring er fortsatt tilgjengelig.',
        transitioning: 'Enhetsstyring: bytter',
        transitioningDescription: 'Enhetsstyringen synkroniseres. Vent litt.',
        grant: 'Gi styring',
        release: 'Frigi',
        releasing: 'Frigir...',
        switching: 'Bytter...',
        releasingLabel: 'Enhetsstyring: frigir',
        releasingDescription: 'Enhetsstyringen gis tilbake. PicoClaw har stoppet pågående inndata.',
        granted: 'PicoClaw-styring gitt',
        released: 'PicoClaw-styring frigitt',
        grantFailed: 'Kunne ikke gi PicoClaw styring',
        releaseFailed: 'Kunne ikke frigi PicoClaw styring',
        grantConfirmTitle: 'Bytte enhetsstyring til PicoClaw?',
        grantConfirmDesc: 'Eksterne MCP-enhetsskrivinger blir avbrutt.'
      },
      install: {
        install: 'Installer PicoClaw',
        installing: 'Installerer PicoClaw',
        success: 'PicoClaw installert vellykket',
        failed: 'Kunne ikke installere PicoClaw',
        uninstalling: 'Avinstallerer runtime...',
        uninstalled: 'Runtime ble avinstallert.',
        uninstallFailed: 'Avinstallering mislyktes.',
        requiredTitle: 'PicoClaw er ikke installert',
        requiredDescription: 'Installer PicoClaw før du starter PicoClaw runtime.',
        progressDescription: 'PicoClaw blir lastet ned og installert.',
        stages: {
          preparing: 'Forbereder',
          downloading: 'Laster ned',
          extracting: 'Pakker ut',
          verifying: 'Verifiserer',
          installing: 'Installerer',
          installed: 'Installert',
          install_timeout: 'Tidsavbrudd',
          install_failed: 'Mislyktes'
        }
      },
      model: {
        requiredTitle: 'Modellkonfigurasjon er nødvendig',
        requiredDescription: 'Konfigurer PicoClaw-modellen før du bruker PicoClaw chat.',
        docsTitle: 'Konfigurasjonsveiledning',
        docsDesc: 'Støttede modeller og protokoller',
        menuLabel: 'Konfigurer modell',
        modelIdentifier: 'Modellidentifikator',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API-nøkkel',
        apiKeyPlaceholder: 'Skriv inn modellens API-nøkkel',
        save: 'Lagre',
        saving: 'Lagrer',
        saved: 'Modellkonfigurasjon lagret',
        saveFailed: 'Kunne ikke lagre modellkonfigurasjonen',
        invalid: 'Modellidentifikator, API Base URL og API-nøkkel kreves'
      },
      uninstall: {
        menuLabel: 'Avinstaller',
        confirmTitle: 'Avinstaller PicoClaw',
        confirmContent:
          'Er du sikker på at du vil avinstallere PicoClaw? Dette vil slette den kjørbare filen og alle konfigurasjonsfilene.',
        confirmOk: 'Avinstaller',
        confirmCancel: 'Avbryt'
      },
      history: {
        title: 'Historikk',
        loading: 'Laster inn økter...',
        emptyTitle: 'Ingen historikk ennå',
        emptyDescription: 'Tidligere PicoClaw økter vil vises her.',
        loadFailed: 'Kunne ikke laste inn økthistorikk',
        deleteFailed: 'Kunne ikke slette økten',
        deleteConfirmTitle: 'Slett økt',
        deleteConfirmContent: 'Er du sikker på at du vil slette "{{title}}"?',
        deleteConfirmOk: 'Slett',
        deleteConfirmCancel: 'Avbryt',
        messageCount_one: '{{count}} melding',
        messageCount_other: '{{count}} meldinger',
        messageCount: '{{count}} meldinger'
      },
      config: {
        startRuntime: 'Start PicoClaw',
        stopRuntime: 'Stopp PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Bytte styringen til PicoClaw?',
        enableConfirmDesc: 'Når PicoClaw startes, deaktiveres den eksterne MCP-tjenesten.',
        enableConfirmOk: 'Start PicoClaw',
        enableConfirmCancel: 'Avbryt',
        title: 'Start PicoClaw',
        description: 'Start runtime for å begynne å bruke PicoClaw-assistenten.',
        switchFromMCP: 'Bytt til PicoClaw og start',
        takeoverAndStart: 'Ta over og start'
      }
    },
    error: {
      title: 'Vi har hatt et problem',
      refresh: 'Oppdater',
      panel: 'Denne delen av siden sluttet å virke',
      retry: 'Prøv igjen'
    },
    fullscreen: {
      toggle: 'Veksle fullskjerm'
    },
    input: {
      disconnected: 'Tastatur og mus er ikke tilkoblet',
      disconnectedTls:
        'Nettleseren avviste den sikre tilkoblingen som overfører tastatur og mus, og det gjør den uten å spørre. Sertifikatet denne enheten genererte er ikke klarert ennå. Åpne denne adressen i en ny fane, godta sertifikatet og last inn siden på nytt. Å installere sertifikatet er den pålitelige løsningen.',
      disconnectedNever:
        'Tilkoblingen som overfører tastatur og mus kunne ikke åpnes. Resten av siden virker fordi den ikke bruker den. Sjekk at ingenting mellom deg og enheten blokkerer den.',
      disconnectedDropped:
        'Tilkoblingen som overfører tastatur og mus ble brutt og har ikke kommet tilbake. Den kobler til igjen av seg selv etter en omstart; hvis dette vedvarer, last inn siden på nytt.',
      hidDisabled: 'HID er slått av på denne enheten (/boot/disable_hid).',
      keyFailed: 'Tasten kunne ikke sendes.'
    },
    speaker: {
      title: 'Høyttaler',
      unmute: 'Slå på lyd',
      mute: 'Demp',
      hostIdle: 'Verten sender ikke lyd',
      hostIdleHint: 'Spill av noe på verten, eller velg KVM-en som lydutgang.'
    },
    upstream: {
      check: 'Se etter oppdateringer',
      updateTo: 'Oppdater til {{version}}',
      confirm: 'Oppdatere {{name}} til {{version}}?',
      confirmDesc:
        'Den nye utgivelsen lastes ned fra GitHub og kontrolleres mot sjekksummene den publiserer. Hvis noe feiler, beholdes nåværende versjon.',
      ok: 'Oppdater',
      upToDate: 'Oppdatert',
      builtIn: 'innebygd',
      checkFailed: 'Kunne ikke se etter oppdateringer: {{error}}',
      unverifiable: 'Versjon {{version}} tilbys ikke: {{reason}}',
      inUse: 'Kan ikke oppdatere nå: {{reason}}',
      running: 'Oppdaterer til {{version}}...',
      done: '{{name}} oppdatert til {{version}}',
      failed: 'Den siste oppdateringen mislyktes: {{error}}'
    },
    menu: {
      mediaAdd: 'Legg til avbildning',
      mediaMoreOptions: 'Flere valg',
      mediaSettings: 'Medieinnstillinger',
      collapse: 'Skjul meny',
      expand: 'Utvid menyen',
      more: 'Mer',
      media: 'Medier',
      tools: 'Verktøy',
      text: 'Tekst',
      advanced: 'Avansert',
      mediaMounted: 'Montert',
      mediaLibrary: 'Bibliotek',
      textToHost: 'Til verten',
      textFromHost: 'Fra verten'
    },
    ion: {
      checking: 'Sjekker videominnet før strømmen starter...',
      warn: 'Lite videominne. Én omstart av serveren ville bruke det opp. Start på nytt når det passer.',
      criticalTitle: 'Ikke nok videominne til å starte strømmen',
      criticalBody:
        'Å starte video ville bruke opp det reserverte minnet og stoppe serveren. Alle andre funksjoner virker fortsatt, også strømstyring og omstart. Bare en omstart av IronKVM frigjør dette minnet.',
      criticalContinue: 'Start video likevel',
      criticalReboot: 'Start IronKVM på nytt',
      criticalRebooting: 'Starter på nytt...'
    }
  }
};

export default nb;
