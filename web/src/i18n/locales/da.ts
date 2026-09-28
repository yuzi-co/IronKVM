const da = {
  translation: {
    head: {
      desktop: 'Fjernskrivebord',
      login: 'Log ind',
      changePassword: 'Skift adgangskode',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        'Browseren nægtede at gemme sessionen. En cookie fra en tidligere HTTPS-session kan ikke erstattes via ukrypteret http. Ryd cookies for denne adresse, eller åbn et privat vindue, og log ind igen.',
      login: 'Log ind',
      placeholderUsername: 'Indtast brugernavn',
      placeholderPassword: 'indtast adgangskode',
      placeholderCurrentPassword: 'Nuværende adgangskode',
      placeholderPassword2: 'indtast adgangskode igen',
      noEmptyUsername: 'brugernavn kan ikke være tom',
      noEmptyPassword: 'adgangskode kan ikke være tom',
      passwordLength: 'Adgangskoden skal være mellem 8 og 72 tegn',
      noAccount:
        'Kunne ikke hente brugeroplysninger. Prøv at opdater siden eller nulstil adgangskoden',
      invalidUser: 'ugyldigt brugernavn eller adgangskode',
      locked: 'For mange logins, prøv venligst igen senere',
      globalLocked: 'System under beskyttelse, prøv venligst igen senere',
      error: 'uventet fejl',
      invalidCurrentPassword: 'Den nuværende adgangskode er forkert',
      changePassword: 'Skift adgangskode',
      changePasswordDesc: 'For sikkerheden af din enhed, bedes du ændre web-login adgangskoden.',
      differentPassword: 'Adgangskoder er ikke ens',
      illegalUsername: 'brugernavn indeholder ugyldige tegn',
      illegalPassword: 'adgangskode indeholder ugyldige tegn',
      forgetPassword: 'Glem adgangskode',
      ok: 'OK',
      cancel: 'Annuller',
      loginButtonText: 'Log ind',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the NanoKVM for 10 seconds.',
        reset2: 'Se detaljerede trin i dette dokument:',
        reset3: 'Standard webkonto:',
        reset4: 'Standard SSH-konto:',
        change1: 'Bemærk, at denne handling ændrer følgende adgangskoder:',
        change2: 'Adgangskode til weblogin',
        change3: 'Systemets root-adgangskode (SSH-loginadgangskode)',
        change4: 'For at nulstille adgangskoderne skal du holde BOOT-knappen på NanoKVM nede.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Konfigurer Wi-Fi for NanoKVM',
      success: 'Please check the network status of NanoKVM and visit the new IP address.',
      failed: 'Handlingen mislykkedes, prøv igen.',
      invalidMode:
        'Den aktuelle tilstand understøtter ikke netværksopsætning. Gå til din enhed og aktiver Wi-Fi konfigurationstilstand.',
      confirmBtn: 'Ok',
      finishBtn: 'Færdig',
      ap: {
        authTitle: 'Godkendelse påkrævet',
        authDescription: 'Indtast venligst AP adgangskoden for at fortsætte',
        authFailed: 'Ugyldig AP adgangskode',
        passPlaceholder: 'AP adgangskode',
        verifyBtn: 'Bekræft'
      }
    },
    screen: {
      scale: 'Skala',
      title: 'Skærm',
      video: 'Videotilstand',
      videoDirectTips: 'Aktiver HTTPS i "Indstillinger > Enhed" for at bruge denne tilstand',
      resolution: 'Opløsning',
      ocr: {
        title: 'Læs tekst (OCR)',
        tips: 'Teksten genkendes i denne browser. Du kan rette den, før du kopierer den.',
        hint: 'Træk hen over den tekst, der skal læses. Tryk på Esc for at annullere.',
        noPicture: 'Vent på videoen, og træk derefter hen over den tekst, der skal læses.',
        cancel: 'Annuller',
        language: 'Sprog',
        languages: {
          eng: 'Engelsk'
        },
        preview: 'Valgt område',
        capturing: 'Optager skærmen...',
        loading: 'Indlæser tekstgenkendelse...',
        recognizing: 'Læser teksten...',
        noText: 'Der blev ikke fundet nogen tekst i det valgte område.',
        copy: 'Kopiér',
        copied: 'Kopieret til udklipsholderen',
        copyFailed: 'Kunne ikke kopiere til udklipsholderen',
        selectAgain: 'Vælg igen',
        unsupported:
          'Denne browser kan ikke køre tekstgenkendelse. Den kræver WebAssembly SIMD, som nuværende browsere har.',
        captureFailed: 'Kunne ikke optage skærmen.',
        outside: 'Det valgte område ligger uden for billedet.',
        recognizeFailed: 'Tekstgenkendelsen mislykkedes.'
      },
      controlRegion: {
        title: 'Musekalibrering',
        description:
          'Brug denne indstilling, når den styrede enhed bruger en opløsning, der ikke er 16:9, og markøren er forskudt vandret eller lodret.',
        off: 'Fra',
        auto: 'Automatisk',
        autoWarning:
          'Kalibreringen kan mislykkes, hvis brugerprogrammet har en helt sort baggrund.',
        manual: 'Manuel',
        selectedResolution: 'Valgt områdeopløsning',
        unused: 'Ikke i brug',
        originalResolution: 'Oprindelig opløsning',
        selectResolution: 'Vælg oprindelig opløsning',
        addResolution: 'Tilføj brugerdefineret opløsning',
        add: 'Tilføj',
        duplicateResolution: 'Denne opløsning findes allerede.',
        width: 'Bredde',
        height: 'Højde',
        apply: 'Beregn og anvend',
        invalidResolution: 'Indtast en gyldig oprindelig opløsning, når videoen er klar.',
        select: 'Vælg område',
        clear: 'Gendan automatisk registrering',
        saveFailed: 'Inputområdet kunne ikke gemmes.',
        tooSmall: 'Det valgte område er for lille.',
        previewUnavailable: 'Forhåndsvisning er ikke tilgængelig',
        clearConfirm: 'Gendan automatisk registrering af sorte kanter?',
        dragHint: 'Træk for at vælge fjernskrivebordets område',
        finish: 'Færdig',
        confirm: 'Bekræft',
        cancel: 'Annuller'
      },
      auto: 'Automatisk',
      autoTips:
        'Screen-tearing eller mouse-offset kan opstå ved enkelte opløsninger. Hvis du oplever dette, kan du prøve at justere fjerncomputerens skærmopløsning eller deaktivere automatisk tilstand.',
      fps: 'FPS',
      customizeFps: 'Tilpas',
      quality: 'Kvalitet',
      qualityLossless: 'Tabsfri',
      qualityHigh: 'Høj',
      qualityMedium: 'Mellem',
      qualityLow: 'Lav',
      frameDetect: 'Beregn frames',
      frameDetectTip:
        'Beregner forskellen mellem hver frame. Stopper med at sende et video stream hvis der ikke registreres ændringer på fjerncomputerens skærm.',
      resetHdmi: 'Nulstil HDMI',
      mixedH264: {
        title: 'H.264-streamkonflikt',
        description:
          'H.264 Direct og H.264 WebRTC bruges samtidigt. Dette kan forårsage skærmrivning eller beskadiget video. Brug kun én H.264-tilstand.'
      },
      webrtcConnectionFailed: {
        title: 'WebRTC-forbindelse mislykkedes',
        description: 'Kontrollér netværksforbindelsen, eller skift videotilstand.'
      },
      captureStatus: {
        hdmiError: 'Fejl i HDMI-billedet',
        unsupportedResolution: 'Den aktuelle opløsning understøttes ikke',
        retrieving: 'Henter skærmbillede...',
        changingResolution: 'Skifter opløsning...',
        updateFailed: 'Skærmbilledet kan ikke opdateres lige nu',
        videoError: 'Fejl i videovisning',
        noHdmi: 'Intet HDMI-signal registreret',
        unavailable: 'Skærmbilledet kan ikke vises lige nu'
      }
    },
    keyboard: {
      title: 'Tastatur',
      paste: 'Indsæt',
      tips: 'Skriver teksten på værten som tastetryk. Vælg det tastaturlayout, værten bruger.',
      placeholder: 'Indtast tekst',
      submit: 'Send',
      virtual: 'Tastatur',
      readClipboard: 'Læs fra udklipsholder',
      clipboardPermissionDenied:
        'Udklipsholdertilladelse nægtet. Tillad venligst udklipsholderadgang i din browser.',
      clipboardReadError: 'Kunne ikke læse udklipsholderen',
      mediaKeys: {
        title: 'Medietaster',
        mute: 'Slå lyden fra',
        volumeDown: 'Skru ned',
        volumeUp: 'Skru op',
        previous: 'Forrige nummer',
        playPause: 'Afspil eller pause',
        next: 'Næste nummer',
        stop: 'Stop'
      },
      pasting: {
        layout: 'Tastaturlayout på værten',
        layouts: {
          us: 'Engelsk (USA)',
          uk: 'Engelsk (Storbritannien)',
          de: 'Tysk',
          fr: 'Fransk',
          es: 'Spansk',
          it: 'Italiensk',
          ptBr: 'Portugisisk (Brasilien)',
          se: 'Svensk / finsk',
          ru: 'Russisk',
          ja: 'Japansk',
          ko: 'Koreansk'
        },
        speed: 'Skrivehastighed',
        speeds: {
          fast: 'Hurtig',
          normal: 'Normal',
          slow: 'Langsom'
        },
        estimate: 'Skrivetid: cirka {{duration}}',
        untypeable: 'Tegn, som dette layout ikke kan skrive: {{count}}',
        untypeableAt: 'linje {{line}}, kolonne {{column}}',
        skipUntypeable: 'Skriv resten',
        shortcut: '{{shortcut}} skriver udklipsholderen på værten med det samme.',
        clipboardUnavailable:
          'Browseren lader kun en side læse udklipsholderen over HTTPS. Indsæt teksten i feltet med Ctrl+V.',
        clipboardEmpty: 'Udklipsholderen indeholder ingen tekst.',
        tooLong: 'Teksten er for lang. Grænsen er {{max}} tegn.',
        inProgress: 'En indsættelse bliver allerede skrevet.',
        typing: 'Skriver på værten',
        done: 'Tekst skrevet',
        canceled: 'Indsættelse annulleret',
        failed: 'Indsættelse mislykkedes',
        cancel: 'Annuller',
        controlBusy: 'En anden styring bruger tastaturet.',
        hidError: 'Tastetrykkene kunne ikke sendes til værten.'
      },
      shortcut: {
        title: 'Genveje',
        custom: 'Brugerdefineret',
        capture: 'Klik her for at fange genvej',
        clear: 'Ryd',
        save: 'Gem',
        captureTips:
          'Optagelse af systemtaster (såsom Windows-tasten) kræver fuldskærmstilladelse.',
        enterFullScreen: 'Skift fuldskærmstilstand.'
      },
      leaderKey: {
        title: 'Leader-tast',
        desc: 'Omgå browserbegrænsninger og send systemgenveje direkte til fjernværten.',
        howToUse: 'Sådan bruges',
        simultaneous: {
          title: 'Samtidig tilstand',
          desc1: 'Hold Leader-tasten nede, og tryk derefter på genvejen.',
          desc2: 'Intuitivt, men kan være i konflikt med systemgenveje.'
        },
        sequential: {
          title: 'Sekventiel tilstand',
          desc1:
            'Tryk på Leader-tasten → tryk på genvejen i rækkefølge → tryk på Leader-tasten igen.',
          desc2: 'Kræver flere trin, men undgår fuldstændig systemkonflikter.'
        },
        enable: 'Aktiver Leader-tast',
        tip: 'Når denne tast tildeles som Leader-tast, fungerer den kun som genvejsudløser og mister sin standardfunktion.',
        placeholder: 'Tryk på Leader-tasten',
        shiftRight: 'Højre Shift',
        ctrlRight: 'Højre Ctrl',
        metaRight: 'Højre Win',
        submit: 'Send',
        recorder: {
          rec: 'REC',
          activate: 'Aktiver taster',
          input: 'Tryk på genvejen...'
        }
      }
    },
    mouse: {
      title: 'Mus',
      cursor: 'Markørstil',
      default: 'Standard-markør',
      pointer: 'Peger-markør',
      cell: 'Celle-markør',
      text: 'Tekst-markør',
      grab: 'Grib-markør',
      hide: 'Skjul mus',
      mode: 'Tilstand for mus',
      absolute: 'Absolut tilstand',
      relative: 'Relativ tilstand',
      absoluteShort: 'Absolut',
      relativeShort: 'Relativ',
      touch: 'Berøringstilstand',
      touchShort: 'Berøring',
      absoluteStalled: 'Målenheden ignorerer den absolutte mus',
      absoluteStalledDesc:
        'Målenheden er holdt op med at modtage absolutte muserapporter, så markørbevægelser går tabt. Tastaturet er ikke påvirket. Gendannelse af USB løser det ofte; relativ tilstand bruger et andet endpoint.',
      useRelative: 'Skift til relativ tilstand',
      direction: 'Rullehjulsretning',
      scrollUp: 'Rul op',
      scrollDown: 'Rul ned',
      speed: 'Rullehjulshastighed',
      fast: 'Hurtigt',
      slow: 'Langsomt',
      requestPointer: 'Bruger relativ-tilstand. Klik på skrivebordet for at få musemarkør.',
      resetHid: 'Nulstil HID',
      hidOnly: {
        title: 'Kun HID-tilstand',
        desc: 'Hvis din mus og tastatur holder op med at reagere, og nulstilling af HID ikke hjælper, kan det være et kompatibilitetsproblem mellem NanoKVM og enheden. Prøv at aktivere HID-Only-tilstand for bedre kompatibilitet.',
        tip1: 'Aktivering af HID-Only-tilstand vil afmontere den virtuelle U-disk og det virtuelle netværk',
        tip2: 'I HID-Only-tilstand er billedmontering deaktiveret',
        rebuild: 'Skift af tilstand genopbygger USB-forbindelsen. NanoKVM genstarter ikke',
        enable: 'Aktiver HID-kun tilstand',
        disable: 'Deaktiver HID-kun tilstand'
      }
    },
    image: {
      title: 'Diskbilleder',
      loading: 'Kontrollerer...',
      empty: 'Ingen fundet',
      mountMode: 'Monteringstilstand',
      mountFailed: 'Montering af diskbillede mislykkedes',
      mountDesc:
        'På nogle systemer kan det være nødvendigt at skubbe den virtuelle disk ud på fjerncomputeren før du kan montere diskbilledet.',
      unmountFailed: 'Afmontering mislykkedes',
      unmountDesc:
        'På nogle systemer skal du manuelt skubbe ud fra fjernværten, før du afmonterer billedet.',
      refresh: 'Opdater billedlisten',
      disk: 'Disk',
      cdrom: 'CD',
      driveEmpty: 'Tom',
      eject: 'Skub ud',
      readOnly: 'Skrivebeskyttet',
      readOnlyTip: 'Gælder for det næste diskbillede, der indsættes i drevet.',
      noDrives: 'Ingen virtuelle drev. Slå den virtuelle disk til under Indstillinger.',
      insertFailed: 'Indsættelse mislykkedes',
      ejectFailed: 'Udskubning mislykkedes',
      insertInto: 'Indsæt i {{drive}}. Klik for at ændre.',
      loadedIn: 'I drevet {{drive}}',
      attention: 'Opmærksomhed påkrævet',
      deleteConfirm: 'Er du sikker på, at du vil slette dette billede?',
      okBtn: 'Ja',
      cancelBtn: 'Annuller',
      deleteFailed: 'Sletning mislykkedes',
      ventoy: {
        statusNoKernel: 'Understøttes ikke af denne firmware',
        statusNotInstalled: 'Ikke installeret',
        statusReady: 'Klar',
        statusSelected: 'Valgte images: {{count}}',
        statusInDrive: 'I diskdrevet, {{size}}',
        noKernel:
          'Denne firmwares kerne har ingen device-mapper-understøttelse, så Ventoy kan ikke bruges, før et image med det er installeret.',
        installDesc: 'Start værten fra flere images på én disk uden at kopiere dem.',
        install: 'Installer',
        installing: 'Henter Ventoy, cirka 20 MB. Det kan tage et par minutter.',
        needsData: 'Ventoy kræver et IronKVM-image med /data-partitionen monteret.',
        uninstall: 'Afinstaller',
        uninstallConfirm: 'Fjern Ventoy-filerne?',
        noImages: 'Ingen images at lægge på Ventoy-disken.',
        onDisk: 'På Ventoy-disken',
        missing: 'Mangler: {{file}}',
        remove: 'Fjern fra Ventoy-disken',
        setHint: 'Sættet af images kan kun ændres, mens Ventoy-disken ikke er i et drev.',
        useAsDisk: 'Brug som virtuel disk',
        failed: 'Ventoy-anmodning mislykkedes',
        secureBoot:
          'Med Secure Boot slået til skal værten én gang registrere Ventoys nøgle i MokManager. Nøglefilen ENROLL_THIS_KEY_IN_MOKMANAGER.cer ligger på VTOYEFI-partitionen.',
        readOnly:
          'Værten ser disken som skrivebeskyttet, så Ventoy-persistens og ventoy.json på drevet virker ikke.'
      },
      tips: {
        title: 'Sådan uploader du',
        usb1: 'Forbind din NanoKVM til din computer via USB.',
        usb2: 'Sørg for, at den virtuelle disk er monteret (Indstillinger -> Virtuel disk).',
        usb3: 'Åben den virtuelle disk på din computer og kopier diskbilledet til roden af den virtuelle disk.',
        scp1: 'Kontroller at din NanoKVM og din computer er på samme lokale netværk.',
        scp2: 'Åben en terminal på din computer og brug SCP-kommandoen for at uploade diskbilledet til /data mappen på din NanoKVM.',
        scp3: 'Eksempel: scp sti-til-dit-diskbillede root@din-nanokvm-ip:/data',
        tfCard: 'microSD-kort',
        tf1: 'Denne metode er understøttet af Linux systemer',
        tf2: 'Tag microSD-kortet ud af din NanoKVM (for den fulde version af NanoKVM skal du åbne enheden for at kunne tage microSD-kortet ud).',
        tf3: 'Indsæt microSD-kortet i en kortlæser og tilslut den til en computer.',
        tf4: 'Kopier diskbilledet til /data mappen på microSD-kortet.',
        tf5: 'Skub microSD-kortet ud og indsæt microSD-kortet i din NanoKVM.'
      }
    },
    script: {
      title: 'Script',
      upload: 'Upload',
      run: 'Kør',
      runBackground: 'Kør i baggrunden',
      runFailed: 'Kørsel mislykkedes',
      attention: 'Opmærksomhed påkrævet',
      delDesc: 'Er du sikker på at du vil slette denne fil?',
      confirm: 'Ja',
      cancel: 'Annuller',
      delete: 'Slet',
      close: 'Luk'
    },
    terminal: {
      title: 'Terminal',
      nanokvm: 'Terminal til NanoKVM',
      serial: 'Terminal til seriel port',
      serialPort: 'Serial port',
      serialPortPlaceholder: 'Angiv seriel port',
      baudrate: 'Baud-hastighed',
      parity: 'Paritet',
      parityNone: 'Ingen',
      parityEven: 'Lige',
      parityOdd: 'Ulige',
      flowControl: 'Flowkontrol',
      flowControlNone: 'Ingen',
      flowControlSoft: 'Software',
      flowControlHard: 'Hardware',
      dataBits: 'Databits',
      stopBits: 'Stopbit',
      confirm: 'OK'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'Sender Wake-on-LAN magic packet',
      sent: 'Wake-on-LAN magic packet sendt',
      input: 'Angiv MAC-adresse',
      ok: 'OK'
    },
    download: {
      title: 'Billedhenter',
      input: 'Indtast venligst et fjernbillede URL',
      ok: 'OK',
      disabled: '/data partitionen er RO, så vi kan ikke downloade billedet',
      uploadbox: 'Slip filen her, eller klik for at vælge',
      inputfile: 'Indtast venligst billedfilen',
      NoISO: 'Ingen ISO',
      sha256: 'SHA-256 (valgfri)',
      sha256Placeholder: 'Indtast en SHA-256-kontrolsum på 64 tegn',
      invalidSHA256: 'SHA-256 skal være en hexadecimal streng på 64 tegn',
      failed: 'Download mislykkedes',
      success: 'Download gennemført',
      checksumFailed: 'Download mislykkedes: SHA-256-verifikation mislykkedes',
      cancel: 'Annuller',
      cancelFailed: 'Kunne ikke annullere download',
      bootMenu: 'Bootmenu (netboot.xyz)',
      bootMenuDesc: "Hent netboot.xyz-ISO'en, med kontrolleret checksum, til den virtuelle cd"
    },
    power: {
      title: 'Tænd/sluk-knap',
      showConfirm: 'Bekræftelse',
      showConfirmTip: 'Strømdrift kræver en ekstra bekræftelse',
      reset: 'Nulstillingsknap',
      power: 'Tænd/sluk-knap',
      powerShort: 'Tænd/sluk-knap (kort tryk)',
      powerLong: 'Tænd/sluk-knap (langt tryk)',
      resetConfirm: 'Fortsæt med nulstilling?',
      powerConfirm: 'Fortsæt strømdrift?',
      okBtn: 'Ja',
      cancelBtn: 'Annuller',
      hostOs: 'Værtens OS',
      hostOsTip: 'Sendes som USB-taster. Værten bestemmer, hvad de gør.',
      sleep: 'Dvale',
      wake: 'Væk',
      wakeKey: 'Væk med Shift',
      powerDown: 'Sluk',
      sleepConfirm: 'Sæt værten i dvale?',
      powerDownConfirm: 'Send sluk-tasten til værten?',
      wakeTip:
        'En vært i dvale ignorerer ofte Væk fra den enhed, der satte den i dvale. Væk med Shift trykker på en tast på tastaturet, som flere værter reagerer på.',
      led: 'Strøm-LED',
      ledOn: 'Tændt',
      ledOff: 'Slukket',
      ledUnknown: 'Ukendt',
      ledConnected: 'Strøm-LED tilsluttet',
      ledConnectedTip:
        "Slå kun til, hvis værtens stikben til strøm-LED'en er forbundet til kortet. Uden den er strømtilstanden ukendt.",
      ledConnectedFailed: 'Kunne ikke gemme indstillingen for strøm-LED'
    },
    settings: {
      title: 'Indstillinger',
      mcp: {
        title: 'MCP-tjeneste',
        service: 'MCP-fjernbetjening',
        serviceDesc: 'Tillad betroede MCP-klienter at styre tastatur og mus og tage skærmbilleder',
        securityWarning:
          'Alle med denne API-nøgle kan styre fjernværten og se dens skærm. Brug HTTPS, og aktivér kun tjenesten på netværk, du har tillid til.',
        endpoint: 'Slutpunkt',
        apiKey: 'API-nøgle',
        regenerateConfirmTitle: 'Generér MCP API-nøglen igen?',
        regenerateConfirmDesc: 'Den nuværende nøgle holder straks op med at virke.',
        enableConfirmTitle: 'Aktivér ekstern MCP-styring?',
        enableConfirmDesc:
          'Aktivering af MCP stopper PicoClaw og lukker alle aktive PicoClaw-sessioner.',
        failed: 'MCP-handlingen mislykkedes',
        copyFailed: 'Kopiering mislykkedes. Kopiér manuelt.',
        okBtn: 'Bekræft',
        cancelBtn: 'Annuller'
      },
      redfish: {
        title: 'Redfish',
        service: 'Redfish-tjeneste',
        serviceDesc:
          'DMTF Redfish API til strømstyring, virtuelle medier og status fra værktøjer som redfishtool og Ansible. Når den slås fra, afsluttes alle Redfish-sessioner.',
        endpoint: 'Tjenesterod',
        httpsOn: 'Kortet bruger HTTPS, som de fleste Redfish-værktøjer kræver.',
        httpsOff:
          'Kortet bruger ukrypteret HTTP. De fleste Redfish-værktøjer kræver HTTPS: slå det til under "Indstillinger > Netværk".',
        credentials:
          'Redfish accepterer KVM-kontiene med Basic-godkendelse eller en Redfish-session samt API-nøgler sendt som X-Auth-Token. API-nøgler administreres på siden API-nøgler.',
        powerActions: 'Strømhandlinger',
        powerActionsDesc:
          'De nulstillingstyper, der tilbydes nu. On, ForceOff og GracefulShutdown kræver strømtilstanden, så de tilbydes kun, når "Strøm-LED tilsluttet" er slået til i strømmenuen.',
        sessions: 'Sessioner',
        noSessions: 'Ingen åbne Redfish-sessioner',
        created: 'Oprettet',
        lastUsed: 'Sidst brugt',
        refresh: 'Opdater',
        end: 'Afslut',
        endConfirmTitle: 'Afslut denne Redfish-session?',
        endConfirmDesc: 'Dens token holder straks op med at virke. Klienten skal logge ind igen.',
        failed: 'Redfish-handling mislykkedes',
        copyFailed: 'Kopiering mislykkedes. Kopiér manuelt.',
        okBtn: 'Bekræft',
        cancelBtn: 'Annuller'
      },
      ipmi: {
        title: 'IPMI',
        warning:
          'IPMI-godkendelse er svag af design. Enhver, der kan nå kortet og kender et brugernavn, kan få et hash af brugerens IPMI-adgangskode og forsøge at knække det offline. Brug genererede adgangskoder, slå kun IPMI til på et netværk, du stoler på, og foretræk Redfish over HTTPS, hvor værktøjet understøtter det.',
        service: 'IPMI over LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) på UDP-port 623 til værtens strøm og status. IPMI 1.5 og cipher suite 0 afvises. Når det slås fra, afsluttes alle IPMI-sessioner.',
        example: 'Eksempel',
        copyFailed: 'Kopiering mislykkedes. Kopiér manuelt.',
        ledOn: 'Strømstatus, on, off, soft, cycle og reset er tilgængelige.',
        ledOff:
          '"Strøm-LED tilsluttet" er slået fra i strømmenuen, så strømtilstanden er ukendt. Kun "power reset" virker: status, on, off, soft og cycle afvises.',
        accounts: 'Konti',
        accountsDesc:
          'IPMI logger ind med KVM-kontiene, hver med sin egen IPMI-adgangskode, adskilt fra webadgangskoden. Administratorer får ADMINISTRATOR. Brugere får USER: de kan læse strømtilstanden med "-L USER", men ikke ændre den.',
        passwordSet: 'IPMI-adgangskode angivet',
        passwordNotSet: 'Ingen IPMI-adgangskode: kan ikke logge ind over IPMI',
        nameTooLong: 'Navnet er længere end 16 tegn, hvilket IPMI ikke tillader',
        accountDisabled: 'Kontoen er deaktiveret',
        setPassword: 'Angiv adgangskode',
        changePassword: 'Skift adgangskode',
        remove: 'Fjern',
        removeConfirmTitle: 'Fjern IPMI-adgangskoden for {{user}}?',
        removeConfirmDesc:
          'Kontoen kan ikke længere logge ind over IPMI, og dens IPMI-sessioner afsluttes.',
        passwordTitle: 'IPMI-adgangskode for {{user}}',
        passwordDesc:
          '12 til 20 skrivbare ASCII-tegn, forskellig fra webadgangskoden. IPMI kræver, at kortet gemmer adgangskoden i en form, det kan læse igen, så brug en, der ikke bruges andre steder. Kopiér den, før du gemmer: den vises ikke igen.',
        passwordPlaceholder: 'IPMI-adgangskode',
        generate: 'Generér',
        copy: 'Kopiér',
        save: 'Gem',
        passwordLength: 'Brug 12 til 20 tegn.',
        passwordChars: 'Brug kun skrivbare ASCII-tegn.',
        saved: 'IPMI-adgangskode gemt',
        failed: 'IPMI-handling mislykkedes',
        okBtn: 'Bekræft',
        cancelBtn: 'Annuller'
      },
      vnc: {
        title: 'VNC',
        service: 'VNC-server',
        serviceDesc:
          'Lader en VNC-klient, f.eks. TigerVNC eller Remmina, se og styre værten. Klienten skal understøtte Tight-kodning. Én session ad gangen.',
        credentials:
          'Log ind med en KVM-konto. Forbindelsen krypteres med kortets TLS-certifikat (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'Den TCP-port, som serveren lytter på.',
        maxFps: 'Grænse for billedhastighed',
        maxFpsDesc: 'Det højeste antal billeder i sekundet, som en klient får.',
        vncAuth: 'Simpel VNC-godkendelse',
        vncAuthDesc:
          'Til klienter uden VeNCrypt. Den tjekker en separat VNC-adgangskode i stedet for en konto.',
        vncAuthWarning:
          'Simpel VNC-godkendelse krypterer ikke forbindelsen. Alle på netværksvejen kan se skærmen og tastetrykkene. Brug den kun på et netværk, du stoler på.',
        password: 'VNC-adgangskode',
        passwordSet: 'Der er angivet en adgangskode. Skriv en ny for at ændre den.',
        passwordInvalid: 'VNC-adgangskoden skal være på 6 til 8 tegn.',
        save: 'Gem',
        saved: 'Indstillinger gemt',
        state: 'Status',
        listening: 'Lytter på port {{port}}',
        notListening: 'Lytter ikke',
        noSession: 'Ingen åben session',
        client: 'Klient',
        user: 'Bruger',
        method: 'Godkendelse',
        methodVencrypt: 'Konto over TLS',
        methodVnc: 'VNC-adgangskode',
        since: 'Forbundet siden',
        resolution: 'Opløsning',
        framesSent: 'Sendte billeder',
        lastError: 'Den seneste session sluttede: {{error}}',
        refresh: 'Opdater',
        disconnect: 'Afbryd',
        disconnectConfirmTitle: 'Afslut VNC-sessionen?',
        disconnectConfirmDesc:
          'Klienten afbrydes med det samme, og alle taster og knapper, den holder nede, slippes.',
        failed: 'VNC-handlingen mislykkedes',
        okBtn: 'Bekræft',
        cancelBtn: 'Annuller'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Værts-watchdog',
        serviceDesc:
          'Hvis værten burde køre, og dens billede ikke ændrer sig, eller der ikke er noget HDMI-signal, i hele tidsgrænsen, trykker kortet på reset eller slukker og tænder værten.',
        stillWarning:
          'En vært, hvis skærm går i dvale, eller hvis billede står stille, mens den arbejder, ser ud til at være gået i stå. Slå skærmdvale fra på værten, eller angiv en ping-adresse.',
        ledHint:
          '"Strøm-LED tilsluttet" er slået fra i strømmenuen. Watchdoggen kan ikke se, hvornår værten er slukket, så den behandler værten som altid tændt.',
        timeout: 'Tidsgrænse',
        timeoutDesc: 'Hvor længe værten må være uden livstegn, før watchdoggen griber ind.',
        action: 'Handling',
        actionDesc:
          'Sluk og tænd holder tænd/sluk-knappen inde i 5 sekunder og trykker derefter på den igen.',
        actionReset: 'Reset',
        actionPower: 'Sluk og tænd',
        cooldown: 'Pause',
        cooldownDesc: 'Den korteste tid mellem to handlinger.',
        maxPerHour: 'Handlinger i timen',
        maxPerHourDesc: 'Det største antal handlinger i en time.',
        pingHost: 'Ping-adresse',
        pingHostDesc:
          'Værtens IP-adresse. Et svar tæller som et livstegn. Lad feltet være tomt for ikke at pinge.',
        pingHostInvalid: 'Angiv en IPv4- eller IPv6-adresse.',
        minutes: 'min.',
        save: 'Gem',
        saved: 'Gemt',
        state: 'Detektor',
        status: {
          off: 'Slået fra',
          watching: 'Overvåger',
          hostOff: 'Vært slukket',
          captureOff: 'HDMI-optagelse slået fra',
          cooldown: 'Pause',
          capped: 'Timegrænse nået',
          acting: 'Griber ind'
        },
        signal: 'HDMI-signal',
        yes: 'Ja',
        no: 'Nej',
        led: 'Strøm-LED',
        on: 'Tændt',
        off: 'Slukket',
        ledNotConnected: 'Ikke tilsluttet',
        ping: 'Ping',
        pingNotSet: 'Ikke angivet',
        pingReply: 'Svarer',
        pingNoReply: 'Intet svar',
        lastChange: 'Seneste billedændring',
        never: 'Aldrig',
        actsIn: 'Griber ind om',
        actionsLastHour: 'Handlinger den seneste time',
        duration: '{{minutes}} min. {{seconds}} s',
        log: 'Log',
        noLog: 'Watchdoggen har ikke grebet ind endnu.',
        refresh: 'Opdater',
        reasonFrozen: 'Billedet ændrede sig ikke',
        reasonNoSignal: 'Intet HDMI-signal',
        stuckFor: 'intet livstegn i {{duration}}',
        pressFailed: 'Tryk mislykkedes: {{error}}',
        noScreenshot: 'Intet skærmbillede',
        failed: 'Watchdog-handlingen mislykkedes'
      },
      netboot: {
        title: 'Netværksboot',
        description:
          "Boot værten fra netværket: iPXE og en menu med billederne på KVM'en over USB-netværksforbindelsen, eller netboot.xyz via proxy-DHCP på LAN'et.",
        addon: 'dnsmasq og bootfiler',
        addonDesc:
          'Installeret på /data: dnsmasq fra Alpine, iPXE og netboot.xyz fra deres udgivelser, hver kontrolleret mod sin checksum.',
        install: 'Installer',
        installing: 'Installerer. Det kan tage et par minutter.',
        uninstall: 'Afinstaller',
        uninstallConfirm: 'Slå netværksboot fra og fjern dnsmasq og bootfilerne?',
        needsData: 'Netværksboot kræver et IronKVM-billede med /data-partitionen monteret.',
        usb: 'På USB-netværksforbindelsen',
        usbDesc:
          "Mens USB-netværksforbindelsen er slået til, betjener dnsmasq den i stedet for udhcpd. Værten får sin ene adresse uden router og uden DNS-server, iPXE til sin arkitektur og en menu med ISO-billederne på KVM'en.",
        linkOff: 'USB-netværksforbindelsen er slået fra. Slå den til under Enhed, USB-netværk.',
        menuUrl: 'Menu',
        leases: 'Værtens lease',
        noLeases: 'Ingen endnu',
        netbootxyzNote:
          'netboot.xyz i menuen indlæses fra internettet, som USB-forbindelsen ikke når. Værten skal have internet på en anden netværksport.',
        lan: "Proxy-DHCP på LAN'et",
        lanDesc:
          "Svarer PXE-klienter på LAN'et med netboot.xyz, som derefter henter sin menu fra internettet. Den uddeler aldrig adresser og stiller ikke billederne på KVM'en til rådighed.",
        lanWarning:
          'Alle PXE-klienter på dette LAN får tilbudt netboot.xyz, ikke kun værten. Slå det kun til på et netværk, du selv styrer.',
        lanConfirm: "Slå proxy-DHCP til på LAN'et?",
        lanInterface: 'LAN',
        running: 'Kører',
        stopped: 'Kører ikke',
        images: 'Billeder i menuen',
        noImages: 'Ingen ISO-billeder i billedmappen.',
        boots: 'Seneste boots',
        noBoots: 'Værten har ikke hentet noget endnu.',
        log: 'dnsmasq-log',
        refresh: 'Opdater',
        okBtn: 'Bekræft',
        cancelBtn: 'Annuller',
        failed: 'Netværksboot-handlingen mislykkedes'
      },
      about: {
        title: 'Om NanoKVM',
        information: 'Information',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Program version',
        applicationTip: 'Version af NanoKVM-webapplikationen',
        image: 'Firmware version',
        imageTip: 'Version af NanoKVM-systemimaget',
        kernel: 'Kerneversion',
        kernelTip: 'Udgaven af den Linux-kerne, der kører nu',
        deviceKey: 'Enhedsnøgle',
        videoMemory: 'Videohukommelse',
        videoMemoryTip:
          'Hukommelse reserveret til videooptagelse. Den deles ikke med resten af systemet.',
        videoMemoryGenerations_one: '{{count}} tidligere NanoKVM-session optager videohukommelse',
        videoMemoryGenerations_other:
          '{{count}} tidligere NanoKVM-sessioner optager videohukommelse',
        videoMemoryReboot: 'Genstart for at frigøre den.',
        community: 'Fællesskab',
        hostname: 'Værtsnavn',
        hostnameUpdated: 'Værtsnavn opdateret. Genstart for at anvende.',
        ipType: {
          Wired: 'Kablet',
          Wireless: 'Trådløs',
          Other: 'Andet'
        }
      },
      appearance: {
        title: 'Udseende',
        display: 'Visning',
        language: 'Sprog',
        languageDesc: 'Vælg sproget til grænsefladen',
        webTitle: 'Webtitel',
        webTitleDesc: 'Tilpas websidens titel',
        menuBar: {
          title: 'Menulinje',
          mode: 'Visningstilstand',
          modeDesc: 'Vis menulinje på skærmen',
          modeOff: 'Fra',
          modeAuto: 'Skjul automatisk',
          modeAlways: 'Altid synlig',
          keyboardLedStatus: 'Tastaturlåseindikatorer',
          keyboardLedStatusDesc:
            'Vis Num Lock-, Caps Lock- og Scroll Lock-status for fjerncomputeren',
          icons: 'Undermenuikoner',
          iconsDesc: 'Vis undermenuikoner i menulinjen'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Status for låse på fjernkeyboard',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Til',
        off: 'Fra',
        unknown: 'Ukendt'
      },
      device: {
        title: 'Enhed',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'OLED-lysstyrke',
          brightnessDescription: 'Et lavere niveau får skærmen til at holde længere',
          brightnessLevels: {
            '64': 'Lavest',
            '96': 'Lav',
            '128': 'Middel',
            '160': 'Høj',
            '207': 'Standard',
            '255': 'Maksimum'
          },
          0: 'Aldrig',
          15: '15 sek.',
          30: '30 sek.',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 time'
        },
        ssh: {
          description: 'Aktiver SSH fjernadgang',
          tip: 'Indstil en stærk adgangskode før aktivering (Konto - Skift adgangskode)'
        },
        advanced: 'Avancerede indstillinger',
        cpuFreq: {
          title: 'CPU-frekvens',
          description: 'Indstil den CPU-takt, der bruges ved næste opstart',
          tip: "CPU'en starter ved 850 MHz og er specificeret til 1000 MHz. En ny værdi anvendes ved næste opstart, ikke mens systemet kører. 1000 MHz er inden for specifikationen; temperaturen er et godt stykke inden for grænserne ved begge indstillinger.",
          running: 'Kører: {{mhz}} MHz',
          rebootToApply: 'genstart for at anvende',
          rebootConfirm: 'Genstart nu for at anvende {{mhz}} MHz?'
        },
        swap: {
          title: 'Byt',
          disable: 'Deaktiver',
          description: 'Indstil swap-filstørrelsen',
          tip: 'Aktivering af denne funktion kan forkorte dit SD-korts brugbare levetid!'
        },
        zram: {
          title: 'Komprimeret swap (zram)',
          description: 'Swap i komprimeret RAM i stedet for på SD-kortet',
          tip: 'zram holder swap væk fra SD-kortet, så det giver intet slid. Der er ingen disk-swap bag det: hvis zram bliver fyldt, stopper kernen en proces i stedet for at swappe langsomt. Hukommelsesgrænsen bestemmer, hvor meget RAM zram må bruge.',
          unavailable: 'Kernemodulerne er ikke installeret på denne enhed',
          inactive: 'Slået til, men enheden startede ikke',
          active: 'Aktiv - {{used}} af {{total}}, {{ratio}}x',
          off: 'Fra',
          detail: {
            algorithm: 'Algoritme: {{algorithm}}',
            memory: 'Hukommelse brugt: {{used}} af {{limit}}',
            memoryNoLimit: 'Hukommelse brugt: {{used}}, ingen grænse angivet',
            counters: 'Sider swappet ind {{in}}, ud {{out}} (alle swap-enheder, siden opstart)'
          }
        },
        mouseJiggler: {
          title: 'Mus Jiggler',
          description: 'Forhindrer fjernværten i at sove',
          disable: 'Deaktiver',
          absolute: 'Absolut tilstand',
          relative: 'Relativ tilstand'
        },
        mdns: {
          description: 'Aktiver mDNS opdagelsestjeneste',
          tip: 'Slukker den, hvis den ikke er nødvendig'
        },
        hdmi: {
          description: 'Aktiver HDMI/monitor output',
          idleTimeoutTitle: 'Timeout for inaktiv optagelse',
          idleTimeoutDescription: 'Stop HDMI-optagelse efter en periode uden aktive seere på',
          minutes: 'min'
        },
        autostart: {
          title: 'Indstillinger for autostart scripts',
          description: 'Administrer scripts, der kører automatisk ved systemstart',
          new: 'Ny',
          deleteConfirm: 'Er du sikker på at du vil slette denne fil?',
          yes: 'Ja',
          no: 'Annuller',
          scriptName: 'Autostart scriptnavn',
          scriptContent: 'Autostart scriptindhold',
          settings: 'Indstillinger'
        },
        hidOnly: 'HID-Kun tilstand',
        hidOnlyDesc:
          'Stop med at emulere virtuelle enheder, og behold kun grundlæggende HID kontrol',
        disk: 'Virtuel disk',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Virtuelt netværk',
        networkDesc: 'Monter det virtuelle netværkskort på den eksterne vært',
        usbNetwork: {
          description:
            'En privat netværksforbindelse til den eksterne vært via USB-kablet. Værten får en adresse uden gateway og uden DNS, så den kan ikke nå dit LAN gennem NanoKVM.',
          off: 'Fra',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (til værter uden NCM)',
          rndis: 'RNDIS (tilbydes ikke længere)',
          rndisNote:
            'Denne forbindelse bruger RNDIS, som ikke længere tilbydes. Vælg NCM eller ECM.',
          subnet: 'Undernet',
          subnetDesc:
            'Et privat IPv4-netværk, /24 til /30. NanoKVM tager den første adresse, værten den anden.',
          addresses: 'NanoKVM: {{board}}, vært: {{host}}',
          invalidSubnet: 'Angiv et undernet, f.eks. 172.31.255.0/30.',
          apply: 'Anvend',
          confirm: 'Genforbind USB-enheden?',
          reenumerate:
            'Når du anvender, genopbygges USB-forbindelsen. Værten mister tastatur, mus og virtuel disk i nogle sekunder.'
        },
        audio: 'Virtuel højttaler',
        audioDesc:
          'Giv fjernværten et USB-lydkort, så du kan høre den. Værten skal vælge det som sin lydudgang. Når du skifter, genopbygges USB-forbindelsen.',
        audioNote: 'Lyd er tilgængelig i begge H.264-tilstande (WebRTC og Direct), ikke i MJPEG',
        console: 'Seriel konsol',
        consoleDesc:
          'Giv fjernværten en seriel USB-port, så du kan logge ind på denne NanoKVM, når netværket ikke kan nås',
        consoleTip:
          'Alle, der styrer fjernværten, får en login-prompt på denne NanoKVM. Indstil en stærk adgangskode før aktivering (Konto - Skift adgangskode).',
        endpoints: {
          title: 'USB-endpoints',
          used: '{{used}} af {{total}} brugt',
          cost: 'bruger {{cost}}',
          needs: 'kræver {{cost}}',
          full: 'Ikke nok USB-endpoints. Slå noget andet fra først.',
          inactive:
            'Slået til, men kører ikke: USB-controlleren løb tør for endpoints. Slå en anden enhed fra, så starter denne med det samme.',
          explain:
            'USB-controlleren har et fast antal indgående endpoints, og det er dem, der tælles her. Hvis flere enheder er slået til, end der er plads til, beholdes tastatur og mus, og resten slås fra.',
          error: 'Kunne ikke nå enheden. Prøv igen.',
          fitTogether: 'Disse passer sammen: {{sets}}'
        },
        reboot: 'Genstart',
        rebootDesc: 'Er du sikker på, at du vil genstarte NanoKVM?',
        okBtn: 'Ja',
        cancelBtn: 'Annuller'
      },
      network: {
        title: 'Netværk',
        wifi: {
          title: 'Wi-Fi',
          description: 'Konfigurer Wi-Fi',
          apMode: 'AP-tilstand er aktiveret, opret forbindelse til Wi-Fi ved at scanne QR-koden',
          connect: 'Tilslut Wi-Fi',
          connectDesc1: 'Indtast netværkets SSID og adgangskode',
          connectDesc2: 'Indtast adgangskoden for at tilslutte dette netværk',
          disconnect: 'Er du sikker på, at du vil afbryde netværket?',
          failed: 'Forbindelsen mislykkedes, prøv igen.',
          ssid: 'Navn',
          password: 'Adgangskode',
          joinBtn: 'Tilslut',
          confirmBtn: 'OK',
          cancelBtn: 'Annuller'
        },
        tls: {
          description: 'Aktiver HTTPS-protokol',
          tip: 'Bemærk: Brug af HTTPS kan øge forsinkelsen, især med MJPEG-videotilstand.',
          restarting: 'Enhedens server genstarter, det tager cirka to minutter...',
          waiting: 'Venter på, at enheden svarer igen...',
          waitingHttp: 'Skifter tilbage til http. Genindlæs siden, hvis den ikke åbner af sig selv.'
        },
        ethernet: {
          title: 'IP-adresse',
          description: 'Konfigurer, hvordan NanoKVM får sin adresse på det kablede netværk',
          dhcp: 'DHCP',
          manual: 'Manuel',
          networkDetails: 'Netværksoplysninger',
          interface: 'Grænseflade',
          ipAddress: 'IP-adresse',
          subnetMask: 'Undernetmaske',
          router: 'Router',
          save: 'Anvend',
          invalidAddress: 'Indtast en gyldig IP-adresse',
          invalidMask: 'Indtast en gyldig undernetmaske, for eksempel 255.255.255.0 eller 24',
          invalidRouter: 'Indtast en gyldig routeradresse',
          addressRequired: 'En IP-adresse er påkrævet',
          maskRequired: 'En undernetmaske er påkrævet',
          applyTitle: 'Vil du ændre adressen på NanoKVM?',
          applyWarning:
            'Forbindelsen til denne side går tabt. NanoKVM anvender den nye adresse og venter {{seconds}} sekunder på, at du når den der. At nå den bevarer ændringen. Hvis intet når den, gendanner NanoKVM de tidligere indstillinger.',
          applyConfirm: 'Anvend',
          applyCancel: 'Annuller',
          applyFailed: 'Adressen kunne ikke anvendes',
          trialTitle: 'Venter på bekræftelse',
          trialDhcp: 'NanoKVM beder om en adresse via DHCP.',
          trialStatic: 'NanoKVM er nu på {{address}}.',
          trialInstruction:
            'Åbn NanoKVM på dens nye adresse, og log ind, hvis den beder om det. At nå den der bevarer ændringen. Hvis intet når NanoKVM inden for {{seconds}} sekunder, gendanner den de tidligere indstillinger.',
          trialOpen: 'Åbn den nye adresse',
          trialKeep: 'Behold disse indstillinger',
          trialKept: 'Den nye adresse er gemt',
          trialKeepFailed: 'Indstillingerne kunne ikke beholdes',
          trialGone: 'Ændringen er allerede gendannet. Prøv igen.',
          unsaved: 'Ikke-gemte ændringer'
        },
        dns: {
          title: 'DNS',
          description: 'Konfigurer DNS-servere til NanoKVM',
          mode: 'Tilstand',
          dhcp: 'DHCP',
          manual: 'Manuel',
          add: 'Tilføj DNS',
          save: 'Gem',
          invalid: 'Indtast en gyldig IP-adresse',
          noDhcp: 'Ingen DHCP-DNS er tilgængelig i øjeblikket',
          saved: 'DNS-indstillinger gemt',
          saveFailed: 'DNS-indstillinger kunne ikke gemmes',
          unsaved: 'Ikke-gemte ændringer',
          maxServers: 'Maksimalt {{count}} DNS-servere er tilladt',
          dnsServers: 'DNS-servere',
          dhcpServersDescription: 'DNS-servere hentes automatisk fra DHCP',
          manualServersDescription: 'DNS-servere kan redigeres manuelt',
          networkDetails: 'Netværksdetaljer',
          interface: 'Grænseflade',
          ipAddress: 'IP-adresse',
          subnetMask: 'Undernetmaske',
          router: 'Router',
          none: 'Ingen'
        }
      },
      vpn: {
        loading: 'Indlæser...',
        okBtn: 'Ja',
        cancelBtn: 'Nej',
        restart: 'Genstart {{name}}?',
        stop: 'Stop {{name}}?',
        stopDesc:
          'Dæmonen stopper nu. "Start ved opstart" er en separat kontakt og forbliver uændret.',
        update: 'Opdater {{name}} til {{version}}?',
        updateDesc: 'Dæmonen genstarter, hvis den kører. Login bevares.',
        notInstall: '{{name}} er ikke installeret.',
        install: 'Installer',
        installing: 'Installerer',
        installFailed: 'Installation mislykkedes',
        retry: 'Prøv igen',
        notRunning: '{{name}} kører ikke. Start den for at fortsætte.',
        run: 'Start',
        boot: 'Start ved opstart',
        bootDesc: "Start {{name}}, når KVM'en starter op.",
        enable: 'Aktiver {{name}}',
        control: 'Kontrolserver',
        connected: 'Forbundet',
        disconnected: 'Ikke forbundet',
        deviceName: 'Enhedsnavn',
        deviceIP: 'Enhedens IP',
        account: 'Konto',
        version: 'Version',
        uptime: 'Oppetid',
        peers: 'Peers',
        noPeers: 'Ingen peers endnu.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Hukommelse',
        daemonRss: 'Dæmon',
        group: 'Tilføjelsesgruppe',
        high: 'begrænses over {{size}}',
        max: 'stoppes af kernen over {{size}}',
        noGroup: 'Ingen hukommelsesgruppe til tilføjelser på dette kort.',
        uninstall: 'Afinstaller {{name}}',
        uninstallDesc:
          'Er du sikker på, at du vil afinstallere {{name}}? Login forbliver på kortet.',
        blocked:
          '{{other}} kører eller starter ved opstart. Der kan kun køre ét VPN ad gangen: stop {{other}} og slå først dens start ved opstart fra.',
        swap: {
          title: 'Swap-hukommelse',
          tip: 'Hvis dæmonen mangler hukommelse, så prøv at aktivere swap-hukommelse. Det sætter som standard swap-filen til 256MB, hvilket kan justeres under "Indstillinger > Enhed".'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Opdater siden og prøv igen. Ellers prøv at installere manuelt.',
        download: 'Download',
        package: 'installationspakken',
        unzip: 'og udpak den',
        upTailscale: 'Upload tailscale til NanoKVM-mappen /usr/bin/',
        upTailscaled: 'Upload tailscaled til NanoKVM-mappen /usr/sbin/',
        refresh: 'Opdater sides',
        notLogin:
          'Enheden er ikke tilknyttet en Tailscale-konto endnu. Log ind for at fuldføre tilknytningen til din konto.',
        urlPeriod: 'Denne URL er gyldig i 10 minutter',
        login: 'Log ind',
        loginSuccess: 'Log ind lykkedes',
        logout: 'Log ud',
        logoutDesc: 'Er du sikker på, at du vil logge ud?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Denne enhed har endnu ikke tilsluttet sig et NetBird-netværk. Tilslut med en opsætningsnøgle, eller log ind med SSO.',
        setupKey: 'Opsætningsnøgle',
        setupKeyPlaceholder: 'Indsæt en opsætningsnøgle fra NetBird-dashboardet',
        join: 'Tilslut',
        or: 'eller',
        sso: 'Log ind med SSO',
        urlPeriod: 'Denne URL er gyldig i 10 minutter',
        loginSuccess: 'Log ind lykkedes',
        logout: 'Afregistrer',
        logoutDesc:
          'Afregistrering fjerner denne peer fra din NetBird-konto og sletter dens konfiguration her. For at tilslutte igen skal du bruge en opsætningsnøgle eller et SSO-login, og peeren kan få en ny IP. Fortsæt?'
      },
      update: {
        title: 'Kontroller for opdatering',
        queryFailed: 'Opdateringskontrol mislykkedes',
        updateFailed: 'Opdatering fejlede. Prøv igen.',
        isLatest: 'Du har allerede den nyeste version.',
        available: 'En opdatering er tilgængelig. Vil du installere den?',
        updating: 'Opdatering i gang. Vent venligst...',
        confirm: 'Bekræft',
        cancel: 'Annuller',
        preview: 'Forhåndsvisning af opdateringer',
        previewDesc: 'Få tidlig adgang til nye funktioner og forbedringer',
        previewTip:
          'Vær opmærksom på, at forhåndsvisningsudgivelser kan indeholde fejl eller ufuldstændig funktionalitet!',
        customServer: {
          title: 'Brugerdefineret opdateringsserver',
          desc: 'Søg efter og download onlineopdateringer fra en angivet server',
          invalidUrl:
            'Indtast en gyldig HTTP- eller HTTPS-servermappe uden forespørgsel, fragment eller latest.json.',
          loadFailed: 'Konfigurationen af opdateringsserveren kunne ikke indlæses.',
          saveFailed: 'Konfigurationen af opdateringsserveren kunne ikke gemmes.',
          saved: 'Konfigurationen af opdateringsserveren er gemt.',
          save: 'Gem',
          confirmTitle: 'Vil du bruge en brugerdefineret opdateringsserver?',
          confirmDesc:
            'SHA-512 kontrollerer kun, at pakken stemmer overens med manifestet fra denne server. Det beviser ikke, at pakken er en officiel NanoKVM-udgivelse. En fejlbehæftet eller ondsindet server kan gøre enheden ubrugelig, medføre tab af data eller kompromittere systemet.',
          confirm: 'Brug alligevel',
          useSipeed: 'Brug den officielle Sipeed-server',
          previewDisabled:
            'Forhåndsvisningsopdateringer er ikke tilgængelige, mens en brugerdefineret opdateringsserver er aktiveret.'
        },
        offline: {
          title: 'Offline opdateringer',
          desc: 'Opdatering via lokal installationspakke',
          upload: 'Upload',
          checksumPlaceholder: 'SHA-256-kontrolsum (valgfri)',
          invalidChecksum: 'SHA-256-kontrolsummen skal indeholde 64 hexadecimale tegn.',
          checksumMismatch: 'SHA-256-verificeringen mislykkedes. Pakken kan være beskadiget.',
          invalidName: 'Ugyldigt filnavnsformat. Download venligst fra GitHub-udgivelser.',
          updateFailed: 'Opdatering fejlede. Prøv igen.'
        }
      },
      account: {
        title: 'Konto',
        webAccount: 'Navn på webkonto',
        role: 'Rolle',
        roles: { admin: 'Administrator', user: 'Bruger' },
        password: 'Adgangskode',
        updateBtn: 'Update',
        logoutBtn: 'Log ud',
        logoutDesc: 'Er du sikker på, at du vil logge ud?',
        okBtn: 'Ja',
        cancelBtn: 'Annuller',
        users: {
          title: 'Brugere',
          create: 'Opret bruger',
          enabled: 'Aktiveret',
          disabled: 'Deaktiveret',
          deviceOwner: 'Enhedsejer',
          resetPassword: 'Nulstil adgangskode',
          delete: 'Slet',
          deleteConfirm: 'Slet denne bruger og tilbagekald alle brugerens sessioner?',
          created: 'Bruger oprettet',
          deleted: 'Bruger slettet',
          passwordUpdated: 'Adgangskode opdateret',
          loadFailed: 'Kunne ikke indlæse brugere',
          saveFailed: 'Kunne ikke gemme bruger',
          deleteFailed: 'Kunne ikke slette bruger'
        }
      },
      apiKeys: {
        title: 'API-nøgler',
        description:
          "En nøgle handler som sin ejer, med den brugers rolle. Send den som Authorization: Bearer <key> til metrics og API'et, eller som X-Auth-Token til Redfish.",
        name: 'Navn',
        namePlaceholder: 'Hvad nøglen bruges til, f.eks. prometheus',
        nameRequired: 'Giv nøglen et navn',
        nameTooLong: 'Navnet må højst være 64 tegn',
        unnamed: '(unavngivet)',
        create: 'Opret nøgle',
        created: 'Oprettet',
        owner: 'Ejer',
        empty: 'Ingen API-nøgler',
        newKeyTitle: 'Din nye API-nøgle',
        newKeyWarning:
          'Kopiér nøglen nu. Den gemmes ikke og kan ikke vises igen. Hvis du mister den, så tilbagekald den og opret en ny.',
        copy: 'Kopiér',
        copied: 'Kopieret',
        copyFailed: 'Kopiering mislykkedes. Kopiér manuelt.',
        done: 'Færdig',
        revoke: 'Tilbagekald',
        revokeConfirmTitle: 'Tilbagekald denne API-nøgle?',
        revokeConfirmDesc: 'Alt, der bruger "{{name}}", holder straks op med at virke.',
        revoked: 'API-nøgle tilbagekaldt',
        loadFailed: 'Kunne ikke indlæse API-nøgler',
        createFailed: 'Kunne ikke oprette API-nøgle',
        revokeFailed: 'Kunne ikke tilbagekalde API-nøgle',
        cancelBtn: 'Annuller'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistent',
      empty: 'Åbn panelet og start en opgave for at begynde.',
      inputPlaceholder: 'Beskriv, hvad du vil have PicoClaw til at gøre',
      newConversation: 'Ny samtale',
      processing: 'Behandler...',
      agent: {
        defaultTitle: 'Generel assistent',
        defaultDescription: 'Generel hjælp til chat, søgning og arbejdsområde.',
        kvmTitle: 'Fjernstyring',
        kvmDescription: 'Betjen fjernværten gennem NanoKVM.',
        switched: 'Agentrolle skiftet',
        switchFailed: 'Kunne ikke skifte agentrolle'
      },
      send: 'Send',
      cancel: 'Annuller',
      status: {
        connecting: 'Opretter forbindelse til gateway...',
        connected: 'PicoClaw-session tilsluttet',
        disconnected: 'PicoClaw-session lukket',
        stopped: 'Stopanmodning sendt',
        runtimeStarted: 'PicoClaw runtime startet',
        runtimeStartFailed: 'Kunne ikke starte PicoClaw runtime',
        runtimeStopped: 'PicoClaw runtime stoppet',
        runtimeStopFailed: 'Kunne ikke stoppe PicoClaw runtime',
        controlSwitchedToMCP: 'Styringen er skiftet til den eksterne MCP-tjeneste'
      },
      connection: {
        runtime: {
          checking: 'Kontrol',
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime klar',
          stopped: 'Runtime stoppet',
          blockedByMCP: 'Ekstern MCP-styring er aktiv',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime utilgængelig',
          configError: 'Konfigurationsfejl'
        },
        transport: {
          connecting: 'Tilslutning',
          connected: 'Tilsluttet',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'Tomgang',
          busy: 'Optaget'
        }
      },
      message: {
        toolAction: 'Handling',
        observation: 'Observation',
        screenshot: 'Skærmbillede'
      },
      overlay: {
        locked: 'PicoClaw styrer enheden. Manuel indtastning er sat på pause.'
      },
      control: {
        picoclaw: 'Enhedsstyring: PicoClaw',
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Enhedsstyring: ekstern MCP',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Enhedsstyring: fra',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Giv styring',
        release: 'Frigiv',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'PicoClaw-styring givet',
        released: 'PicoClaw-styring frigivet',
        grantFailed: 'Kunne ikke give PicoClaw styring',
        releaseFailed: 'Kunne ikke frigive PicoClaw styring',
        grantConfirmTitle: 'Skift enhedsstyring til PicoClaw?',
        grantConfirmDesc: 'Eksterne MCP-enhedsskrivninger bliver afbrudt.'
      },
      install: {
        install: 'Installer PicoClaw',
        installing: 'Installation af PicoClaw',
        success: 'PicoClaw installeret korrekt',
        failed: 'Kunne ikke installere PicoClaw',
        uninstalling: 'Afinstallerer runtime...',
        uninstalled: 'Runtime blev afinstalleret.',
        uninstallFailed: 'Afinstallation mislykkedes.',
        requiredTitle: 'PicoClaw er ikke installeret',
        requiredDescription: 'Installer PicoClaw før start af PicoClaw runtime.',
        progressDescription: 'PicoClaw bliver downloadet og installeret.',
        stages: {
          preparing: 'Forberedelse',
          downloading: 'Downloader',
          extracting: 'Udpakning',
          verifying: 'Bekræfter',
          installing: 'Installerer',
          installed: 'Installeret',
          install_timeout: 'Timeout',
          install_failed: 'Mislykkedes'
        }
      },
      model: {
        requiredTitle: 'Modelkonfiguration er påkrævet',
        requiredDescription: 'Konfigurer PicoClaw-modellen, før du bruger PicoClaw chat.',
        docsTitle: 'Konfigurationsvejledning',
        docsDesc: 'Understøttede modeller og protokoller',
        menuLabel: 'Konfigurer model',
        modelIdentifier: 'Modelidentifikator',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API-nøgle',
        apiKeyPlaceholder: 'Indtast modellens API-nøgle',
        save: 'Gem',
        saving: 'Gemmer',
        saved: 'Modelkonfiguration gemt',
        saveFailed: 'Kunne ikke gemme modelkonfigurationen',
        invalid: 'Model-id, API Base URL og API-nøgle er påkrævet'
      },
      uninstall: {
        menuLabel: 'Afinstaller',
        confirmTitle: 'Afinstaller PicoClaw',
        confirmContent:
          'Er du sikker på, at du vil afinstallere PicoClaw? Dette vil slette den eksekverbare fil og alle konfigurationsfiler.',
        confirmOk: 'Afinstaller',
        confirmCancel: 'Annuller'
      },
      history: {
        title: 'Historik',
        loading: 'Indlæser sessioner...',
        emptyTitle: 'Ingen historik endnu',
        emptyDescription: 'Tidligere PicoClaw sessioner vil blive vist her.',
        loadFailed: 'Kunne ikke indlæse sessionshistorikken',
        deleteFailed: 'Kunne ikke slette session',
        deleteConfirmTitle: 'Slet session',
        deleteConfirmContent: 'Er du sikker på, at du vil slette "{{title}}"?',
        deleteConfirmOk: 'Slet',
        deleteConfirmCancel: 'Annuller',
        messageCount_one: '{{count}} besked',
        messageCount_other: '{{count}} beskeder',
        messageCount: '{{count}} beskeder'
      },
      config: {
        startRuntime: 'Start PicoClaw',
        stopRuntime: 'Stop PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Skift styringen til PicoClaw?',
        enableConfirmDesc: 'Start af PicoClaw deaktiverer den eksterne MCP-tjeneste.',
        enableConfirmOk: 'Start PicoClaw',
        enableConfirmCancel: 'Annuller',
        title: 'Start PicoClaw',
        description: 'Start runtime for at begynde at bruge PicoClaw-assistenten.',
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'Vi er stødt på et problem',
      refresh: 'Opdater',
      panel: 'Denne del af siden holdt op med at virke',
      retry: 'Prøv igen'
    },
    fullscreen: {
      toggle: 'Skift fuldskærm'
    },
    input: {
      disconnected: 'Tastatur og mus er ikke forbundet',
      disconnectedTls:
        'Browseren afviste den sikre forbindelse, der bærer tastatur og mus, og det gør den uden at spørge. Certifikatet, som denne enhed har genereret, er endnu ikke betroet. Åbn denne adresse i en ny fane, accepter certifikatet, og genindlæs derefter. Den pålidelige løsning er at installere certifikatet.',
      disconnectedNever:
        'Forbindelsen, der bærer tastatur og mus, kunne ikke åbnes. Resten af siden virker, fordi den ikke bruger forbindelsen. Kontroller, at intet mellem dig og enheden blokerer den.',
      disconnectedDropped:
        'Forbindelsen, der bærer tastatur og mus, blev afbrudt og er ikke kommet tilbage. Den genopretter sig selv efter en genstart; hvis dette bliver ved, så genindlæs siden.',
      hidDisabled: 'HID er slået fra på denne enhed (/boot/disable_hid).',
      keyFailed: 'Tasten kunne ikke sendes.'
    },
    speaker: { title: 'Højttaler', unmute: 'Slå lyd til', mute: 'Slå lyd fra' },
    menu: {
      collapse: 'Skjul menu',
      expand: 'Udvid menu'
    },
    ion: {
      checking: 'Kontrollerer videohukommelsen, før streamen startes...',
      warn: 'Videohukommelsen er lav. Én genstart af serveren ville opbruge den. Genstart, når det passer dig.',
      criticalTitle: 'Ikke nok videohukommelse til at starte streamen',
      criticalBody:
        'Hvis videoen startes, opbruges den reserverede hukommelse, og serveren stopper. Alle andre funktioner virker stadig, herunder strømstyring og genstart. Kun en genstart af NanoKVM frigør denne hukommelse.',
      criticalContinue: 'Start video alligevel',
      criticalReboot: 'Genstart NanoKVM',
      criticalRebooting: 'Genstarter...'
    }
  }
};

export default da;
