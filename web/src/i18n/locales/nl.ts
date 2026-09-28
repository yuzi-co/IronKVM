const nl = {
  translation: {
    head: {
      desktop: 'Extern bureaublad',
      login: 'Inloggen',
      changePassword: 'Wachtwoord wijzigen',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        'De browser weigerde de sessie op te slaan. Een cookie van een eerdere HTTPS-sessie kan niet via onversleuteld http worden vervangen. Wis de cookies voor dit adres of open een privévenster en log opnieuw in.',
      login: 'Inloggen',
      placeholderUsername: 'Voer gebruikersnaam in',
      placeholderPassword: 'Voer wachtwoord in',
      placeholderCurrentPassword: 'Huidig wachtwoord',
      placeholderPassword2: 'Voer wachtwoord nogmaals in',
      noEmptyUsername: 'Gebruikersnaam mag niet leeg zijn',
      noEmptyPassword: 'Wachtwoord mag niet leeg zijn',
      passwordLength: 'Het wachtwoord moet tussen 8 en 72 tekens lang zijn',
      noAccount:
        'Ophalen van gebruikersinformatie mislukt, vernieuw de webpagina of reset het wachtwoord',
      invalidUser: 'Ongeldige gebruikersnaam of wachtwoord',
      locked: 'Te veel aanmeldingen, probeer het later opnieuw',
      globalLocked: 'Systeem wordt beveiligd. Probeer het later opnieuw',
      error: 'Onverwachte fout',
      invalidCurrentPassword: 'Het huidige wachtwoord is onjuist',
      changePassword: 'Wachtwoord wijzigen',
      changePasswordDesc:
        'Voor de veiligheid van uw apparaat, wijzig alstublieft het webaanmeldingswachtwoord.',
      differentPassword: 'Wachtwoorden komen niet overeen',
      illegalUsername: 'Gebruikersnaam bevat ongeldige tekens',
      illegalPassword: 'Wachtwoord bevat ongeldige tekens',
      forgetPassword: 'Wachtwoord vergeten',
      ok: 'Ok',
      cancel: 'Annuleren',
      loginButtonText: 'Inloggen',
      tips: {
        reset1:
          'Om de wachtwoorden opnieuw in te stellen, houdt u de BOOT-knop op de NanoKVM 10 seconden lang ingedrukt.',
        reset2: 'Voor gedetailleerde stappen kunt u dit document raadplegen:',
        reset3: 'Standaard webaccount:',
        reset4: 'Standaard SSH-account:',
        change1: 'Houd er rekening mee dat deze actie de volgende wachtwoorden zal wijzigen:',
        change2: 'Web login wachtwoord',
        change3: 'Systeem root-wachtwoord (SSH-inlogwachtwoord)',
        change4:
          'Om de wachtwoorden opnieuw in te stellen, houdt u de BOOT-knop op de NanoKVM ingedrukt.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Wi-Fi configureren voor NanoKVM',
      success: 'Controleer de netwerkstatus van NanoKVM en bezoek het nieuwe IP-adres.',
      failed: 'De bewerking is mislukt. Probeer het opnieuw.',
      invalidMode:
        'De huidige modus ondersteunt geen netwerkconfiguratie. Ga naar uw apparaat en schakel de configuratiemodus Wi-Fi in.',
      confirmBtn: 'Ok',
      finishBtn: 'Gereed',
      ap: {
        authTitle: 'Authenticatie vereist',
        authDescription: 'Voer het wachtwoord AP in om door te gaan',
        authFailed: 'Ongeldig AP wachtwoord',
        passPlaceholder: 'AP wachtwoord',
        verifyBtn: 'Verifieer'
      }
    },
    screen: {
      scale: 'Schaal',
      title: 'Scherm',
      video: 'Videomodus',
      videoDirectTips: 'Schakel HTTPS in "Instellingen > Apparaat" in om deze modus te gebruiken',
      resolution: 'Resolutie',
      ocr: {
        title: 'Tekst lezen (OCR)',
        tips: 'De tekst wordt in deze browser herkend. U kunt hem verbeteren voordat u hem kopieert.',
        hint: 'Sleep over de tekst die u wilt lezen. Druk op Esc om te annuleren.',
        noPicture: 'Wacht op de video en sleep daarna over de tekst die u wilt lezen.',
        cancel: 'Annuleren',
        language: 'Taal',
        languages: {
          eng: 'Engels'
        },
        preview: 'Geselecteerd gebied',
        capturing: 'Scherm vastleggen...',
        loading: 'Tekstherkenning laden...',
        recognizing: 'Tekst lezen...',
        noText: 'Er is geen tekst gevonden in het geselecteerde gebied.',
        copy: 'Kopiëren',
        copied: 'Gekopieerd naar het klembord',
        copyFailed: 'Kopiëren naar het klembord is mislukt',
        selectAgain: 'Opnieuw selecteren',
        unsupported:
          'Deze browser kan geen tekstherkenning uitvoeren. Daarvoor is WebAssembly SIMD nodig, dat huidige browsers hebben.',
        captureFailed: 'Het scherm kon niet worden vastgelegd.',
        outside: 'Het geselecteerde gebied ligt buiten het beeld.',
        recognizeFailed: 'Tekstherkenning is mislukt.'
      },
      controlRegion: {
        title: 'Muiskalibratie',
        description:
          'Gebruik deze instelling wanneer het bestuurde apparaat een andere resolutie dan 16:9 gebruikt en de cursor horizontaal of verticaal niet goed is uitgelijnd.',
        off: 'Uit',
        auto: 'Automatisch',
        autoWarning:
          'De kalibratie kan mislukken als de gebruikerstoepassing een volledig zwarte achtergrond heeft.',
        manual: 'Handmatig',
        selectedResolution: 'Resolutie van geselecteerd gebied',
        unused: 'Niet gebruikt',
        originalResolution: 'Oorspronkelijke resolutie',
        selectResolution: 'Oorspronkelijke resolutie selecteren',
        addResolution: 'Aangepaste resolutie toevoegen',
        add: 'Toevoegen',
        duplicateResolution: 'Deze resolutie bestaat al.',
        width: 'Breedte',
        height: 'Hoogte',
        apply: 'Berekenen en toepassen',
        invalidResolution:
          'Voer een geldige oorspronkelijke resolutie in zodra de video gereed is.',
        select: 'Gebied selecteren',
        clear: 'Automatische detectie herstellen',
        saveFailed: 'Het invoergebied kan niet worden opgeslagen.',
        tooSmall: 'Het geselecteerde gebied is te klein.',
        previewUnavailable: 'Voorbeeld niet beschikbaar',
        clearConfirm: 'Automatische detectie van zwarte randen herstellen?',
        dragHint: 'Sleep om het externe bureaubladgebied te selecteren',
        finish: 'Gereed',
        confirm: 'Bevestigen',
        cancel: 'Annuleren'
      },
      auto: 'Automatisch',
      autoTips:
        'Bij bepaalde resoluties kunnen schermverscheuringen of muisverplaatsingen optreden. Overweeg de resolutie van de externe host aan te passen of schakel de automatische modus uit.',
      fps: 'FPS',
      customizeFps: 'Aanpassen',
      quality: 'Kwaliteit',
      qualityLossless: 'Verliesvrij',
      qualityHigh: 'Hoog',
      qualityMedium: 'Gemiddeld',
      qualityLow: 'Laag',
      frameDetect: 'Frame detectie',
      frameDetectTip:
        'Berekent het verschil tussen frames. Stopt met het verzenden van de videostream wanneer er geen veranderingen worden gedetecteerd op het scherm van de externe host.',
      resetHdmi: 'Reset HDMI',
      mixedH264: {
        title: 'H.264-streamconflict',
        description:
          'H.264 Direct en H.264 WebRTC worden tegelijkertijd gebruikt. Dit kan tearing of beschadigde video veroorzaken. Gebruik slechts één H.264-modus.'
      },
      webrtcConnectionFailed: {
        title: 'WebRTC-verbinding mislukt',
        description: 'Controleer de netwerkverbinding of wijzig de videomodus.'
      },
      captureStatus: {
        hdmiError: 'HDMI-schermfout',
        unsupportedResolution: 'De huidige resolutie wordt niet ondersteund',
        retrieving: 'Scherm ophalen...',
        changingResolution: 'Resolutie wisselen...',
        updateFailed: 'Het scherm kan nu niet worden bijgewerkt',
        videoError: 'Fout bij videoweergave',
        noHdmi: 'Geen HDMI-signaal gedetecteerd',
        unavailable: 'Het scherm kan nu niet worden weergegeven'
      }
    },
    keyboard: {
      title: 'Toetsenbord',
      paste: 'Plakken',
      tips: 'Typt de tekst op de host als toetsaanslagen. Kies de toetsenbordindeling die de host gebruikt.',
      placeholder: 'Voer tekst in',
      submit: 'Verzenden',
      virtual: 'Toetsenbord',
      readClipboard: 'Lezen vanaf het Klembord',
      clipboardPermissionDenied:
        'Klembordtoestemming geweigerd. Sta klembordtoegang toe in uw browser.',
      clipboardReadError: 'Kan het klembord niet lezen',
      mediaKeys: {
        title: 'Mediatoetsen',
        mute: 'Dempen',
        volumeDown: 'Volume omlaag',
        volumeUp: 'Volume omhoog',
        previous: 'Vorige track',
        playPause: 'Afspelen of pauzeren',
        next: 'Volgende track',
        stop: 'Stoppen'
      },
      pasting: {
        layout: 'Toetsenbordindeling van de host',
        layouts: {
          us: 'Engels (VS)',
          uk: 'Engels (VK)',
          de: 'Duits',
          fr: 'Frans',
          es: 'Spaans',
          it: 'Italiaans',
          ptBr: 'Portugees (Brazilië)',
          se: 'Zweeds / Fins',
          ru: 'Russisch',
          ja: 'Japans',
          ko: 'Koreaans'
        },
        speed: 'Typsnelheid',
        speeds: {
          fast: 'Snel',
          normal: 'Normaal',
          slow: 'Langzaam'
        },
        estimate: 'Typtijd: ongeveer {{duration}}',
        untypeable: 'Tekens die deze indeling niet kan typen: {{count}}',
        untypeableAt: 'regel {{line}}, kolom {{column}}',
        skipUntypeable: 'De rest typen',
        shortcut: '{{shortcut}} typt het klembord meteen op de host.',
        clipboardUnavailable:
          'De browser laat een pagina het klembord alleen via HTTPS lezen. Plak de tekst met Ctrl+V in het vak.',
        clipboardEmpty: 'Het klembord bevat geen tekst.',
        tooLong: 'De tekst is te lang. De limiet is {{max}} tekens.',
        inProgress: 'Er wordt al een geplakte tekst getypt.',
        typing: 'Typen op de host',
        done: 'Tekst getypt',
        canceled: 'Plakken geannuleerd',
        failed: 'Plakken mislukt',
        cancel: 'Annuleren',
        controlBusy: 'Een andere besturing gebruikt het toetsenbord.',
        hidError: 'De toetsaanslagen konden niet naar de host worden gestuurd.'
      },
      shortcut: {
        title: 'Snelkoppelingen',
        custom: 'Aangepast',
        capture: 'Klik hier om de snelkoppeling vast te leggen',
        clear: 'Duidelijk',
        save: 'Opslaan',
        captureTips:
          'Voor het vastleggen van systeemtoetsen (zoals de Windows-toets) is toestemming voor volledig scherm vereist.',
        enterFullScreen: 'Schakelen naar volledig scherm.'
      },
      leaderKey: {
        title: 'Leader-toets',
        desc: 'Omzeil browserbeperkingen en stuur systeemsnelkoppelingen rechtstreeks naar de externe host.',
        howToUse: 'Hoe te gebruiken',
        simultaneous: {
          title: 'Gelijktijdige modus',
          desc1: 'Houd de Leader-toets ingedrukt en druk daarna op de sneltoets.',
          desc2: 'Intuïtief, maar kan conflicteren met systeemsnelkoppelingen.'
        },
        sequential: {
          title: 'Sequentiële modus',
          desc1:
            'Druk op de Leader-toets → druk de sneltoets in volgorde in → druk opnieuw op de Leader-toets.',
          desc2: 'Vereist meer stappen, maar vermijdt systeemconflicten volledig.'
        },
        enable: 'Leader-toets inschakelen',
        tip: 'Wanneer deze toets als Leader-toets is toegewezen, werkt hij uitsluitend als sneltoetstrigger en verliest hij zijn standaardgedrag.',
        placeholder: 'Druk op de Leader-toets',
        shiftRight: 'Rechter Shift',
        ctrlRight: 'Rechter Ctrl',
        metaRight: 'Rechter Win',
        submit: 'Verzenden',
        recorder: {
          rec: 'OPN',
          activate: 'Toetsen activeren',
          input: 'Druk op de snelkoppeling...'
        }
      }
    },
    mouse: {
      title: 'Muis',
      cursor: 'Cursorstijl',
      default: 'Standaard cursor',
      pointer: 'Aanwijzer cursor',
      cell: 'Cel cursor',
      text: 'Tekst cursor',
      grab: 'Grijp cursor',
      hide: 'Cursor verbergen',
      mode: 'Muismodus',
      absolute: 'Absolute modus',
      relative: 'Relatieve modus',
      absoluteShort: 'Absoluut',
      relativeShort: 'Relatief',
      touch: 'Aanraakmodus',
      touchShort: 'Aanraken',
      absoluteStalled: 'Het doelapparaat negeert de absolute muis',
      absoluteStalledDesc:
        'Het doelapparaat neemt geen absolute muisrapporten meer aan, dus aanwijzerbewegingen gaan verloren. Het toetsenbord werkt gewoon. Het herstellen van USB verhelpt dit vaak; de relatieve modus gebruikt een ander endpoint.',
      useRelative: 'Overschakelen naar relatieve modus',
      direction: 'Scrollwielrichting',
      scrollUp: 'Scroll naar boven',
      scrollDown: 'Scroll naar beneden',
      speed: 'Scrollwielsnelheid',
      fast: 'Snel',
      slow: 'Langzaam',
      requestPointer:
        'Relatieve modus wordt gebruikt. Klik op het bureaublad om de muisaanwijzer te krijgen.',
      resetHid: 'HID resetten',
      hidOnly: {
        title: 'Alleen HID-modus',
        desc: 'Als uw muis en toetsenbord niet meer reageren en het opnieuw instellen van HID niet helpt, kan er sprake zijn van een compatibiliteitsprobleem tussen de NanoKVM en het apparaat. Probeer de modus HID-Only in te schakelen voor betere compatibiliteit.',
        tip1: 'Als u de modus HID-Only inschakelt, worden de virtuele U-schijf en het virtuele netwerk ontkoppeld',
        tip2: 'In de modus HID-Alleen is beeldmontage uitgeschakeld',
        rebuild:
          'Bij het wisselen van modus wordt de USB-verbinding opnieuw opgebouwd. NanoKVM start niet opnieuw op',
        enable: 'Schakel de modus HID-Alleen in',
        disable: 'Schakel de modus HID-Alleen uit'
      }
    },
    image: {
      title: 'Afbeeldingen',
      loading: 'Laden...',
      empty: 'Niets gevonden',
      mountMode: 'Montagemodus',
      mountFailed: 'Koppelen mislukt',
      mountDesc:
        'In sommige systemen is het noodzakelijk om de virtuele schijf op de externe host uit te werpen voordat het image wordt gekoppeld.',
      unmountFailed: 'Ontkoppelen mislukt',
      unmountDesc:
        'Op sommige systemen moet u de image handmatig uitwerpen van de externe host voordat u de image ontkoppelt.',
      refresh: 'Vernieuw de afbeeldingenlijst',
      disk: 'Schijf',
      cdrom: 'CD',
      driveEmpty: 'Leeg',
      eject: 'Uitwerpen',
      readOnly: 'Alleen-lezen',
      readOnlyTip: 'Geldt voor de volgende image die in de schijf wordt geplaatst.',
      noDrives: 'Geen virtuele stations. Schakel de virtuele schijf in bij Instellingen.',
      insertFailed: 'Plaatsen mislukt',
      ejectFailed: 'Uitwerpen mislukt',
      insertInto: 'Plaatsen in {{drive}}. Klik om te wijzigen.',
      loadedIn: 'In station {{drive}}',
      attention: 'Let op',
      deleteConfirm: 'Weet u zeker dat u deze afbeelding wilt verwijderen?',
      okBtn: 'Ja',
      cancelBtn: 'Nee',
      deleteFailed: 'Verwijderen mislukt',
      ventoy: {
        statusNoKernel: 'Niet ondersteund door deze firmware',
        statusNotInstalled: 'Niet geïnstalleerd',
        statusReady: 'Gereed',
        statusSelected: 'Geselecteerde images: {{count}}',
        statusInDrive: 'In het schijfstation, {{size}}',
        noKernel:
          'De kernel van deze firmware ondersteunt geen device-mapper, dus Ventoy kan pas worden gebruikt nadat een image met die ondersteuning is geïnstalleerd.',
        installDesc: 'Start de host op vanaf meerdere images op één schijf, zonder ze te kopiëren.',
        install: 'Installeren',
        installing: 'Ventoy wordt gedownload, ongeveer 20 MB. Dit kan enkele minuten duren.',
        needsData: 'Ventoy heeft een IronKVM-image nodig met de /data-partitie gekoppeld.',
        uninstall: 'Verwijderen',
        uninstallConfirm: 'Ventoy-bestanden verwijderen?',
        noImages: 'Geen images voor de Ventoy-schijf.',
        onDisk: 'Op de Ventoy-schijf',
        missing: 'Ontbreekt: {{file}}',
        remove: 'Van de Ventoy-schijf halen',
        setHint:
          'De selectie van images kan alleen veranderen zolang de Ventoy-schijf in geen enkel station zit.',
        useAsDisk: 'Gebruiken als virtuele schijf',
        failed: 'Ventoy-verzoek mislukt',
        secureBoot:
          'Met Secure Boot aan moet de host de sleutel van Ventoy eenmalig registreren in MokManager. Het sleutelbestand ENROLL_THIS_KEY_IN_MOKMANAGER.cer staat op de VTOYEFI-partitie.',
        readOnly:
          'De host ziet de schijf als alleen-lezen, dus Ventoy-persistentie en ventoy.json op het station werken niet.'
      },
      tips: {
        title: 'Hoe te uploaden',
        usb1: 'Verbind de NanoKVM met uw computer via USB.',
        usb2: 'Zorg ervoor dat de virtuele schijf is gekoppeld (Instellingen - Virtuele schijf).',
        usb3: 'Open de virtuele schijf op uw computer en kopieer het imagebestand naar de hoofdmap van de virtuele schijf.',
        scp1: 'Zorg ervoor dat de NanoKVM en uw computer zich in hetzelfde lokale netwerk bevinden.',
        scp2: 'Open een terminal op uw computer en gebruik het SCP-commando om het imagebestand te uploaden naar de /data directory op de NanoKVM.',
        scp3: 'Voorbeeld: scp uw-image-pad root@uw-nanokvm-ip:/data',
        tfCard: 'TF-kaart',
        tf1: 'Deze methode wordt ondersteund op Linux-systemen',
        tf2: 'Haal de TF-kaart uit de NanoKVM (voor de VOLLEDIGE versie, demonteer eerst de behuizing).',
        tf3: 'Plaats de TF-kaart in een kaartlezer en verbind deze met uw computer.',
        tf4: 'Kopieer het imagebestand naar de /data directory op de TF-kaart.',
        tf5: 'Plaats de TF-kaart terug in de NanoKVM.'
      }
    },
    script: {
      title: 'Script',
      upload: 'Uploaden',
      run: 'Uitvoeren',
      runBackground: 'Op achtergrond uitvoeren',
      runFailed: 'Uitvoeren mislukt',
      attention: 'Let op',
      delDesc: 'Weet u zeker dat u dit bestand wilt verwijderen?',
      confirm: 'Ja',
      cancel: 'Nee',
      delete: 'Verwijderen',
      close: 'Sluiten'
    },
    terminal: {
      title: 'Terminal',
      nanokvm: 'NanoKVM Terminal',
      serial: 'Seriële poort terminal',
      serialPort: 'Seriële poort',
      serialPortPlaceholder: 'Voer de seriële poort in',
      baudrate: 'Baudrate',
      parity: 'Pariteit',
      parityNone: 'Geen',
      parityEven: 'Even',
      parityOdd: 'Oneven',
      flowControl: 'Debietregeling',
      flowControlNone: 'Geen',
      flowControlSoft: 'Software',
      flowControlHard: 'Hardware',
      dataBits: 'Databits',
      stopBits: 'Stopbits',
      confirm: 'Ok'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'Commando wordt verzonden...',
      sent: 'Commando verzonden',
      input: 'Voer het MAC-adres in',
      ok: 'Ok'
    },
    download: {
      title: 'Afbeeldingdownloader',
      input: 'Voer een externe afbeelding in URL',
      ok: 'Ok',
      disabled: '/data partitie is RO, dus we kunnen de afbeelding niet downloaden',
      uploadbox: 'Zet het bestand hier neer of klik om te selecteren',
      inputfile: 'Voer het afbeeldingsbestand in',
      NoISO: 'Geen ISO',
      sha256: 'SHA-256 (optioneel)',
      sha256Placeholder: 'Voer een SHA-256-controlesom van 64 tekens in',
      invalidSHA256: 'SHA-256 moet een hexadecimale tekenreeks van 64 tekens zijn',
      failed: 'Download mislukt',
      success: 'Download geslaagd',
      checksumFailed: 'Download mislukt: SHA-256-verificatie mislukt',
      cancel: 'Annuleren',
      cancelFailed: 'Download annuleren mislukt',
      bootMenu: 'Opstartmenu (netboot.xyz)',
      bootMenuDesc:
        'De netboot.xyz-ISO downloaden, met gecontroleerde checksum, voor de virtuele cd'
    },
    power: {
      title: 'Aan/uit',
      showConfirm: 'Bevestiging',
      showConfirmTip:
        'Vragen voor een korte druk op de aan/uit-knop. Reset en lang drukken vragen altijd.',
      reset: 'Resetten',
      power: 'Aan/uit',
      powerShort: 'Aan/uit (kort indrukken)',
      powerLong: 'Aan/uit (lang indrukken)',
      resetConfirm: 'Doorgaan met resetten?',
      powerConfirm: 'Doorgaan met stroomvoorziening?',
      okBtn: 'Ja',
      cancelBtn: 'Nee',
      hostOs: 'Host-OS',
      hostOsTip: 'Verzonden als USB-toetsen. De host bepaalt wat ze doen.',
      sleep: 'Slaapstand',
      wake: 'Wekken',
      wakeKey: 'Wekken met Shift',
      powerDown: 'Uitschakelen',
      sleepConfirm: 'De host in slaapstand zetten?',
      powerDownConfirm: 'De uitschakeltoets naar de host sturen?',
      wakeTip:
        'Een slapende host negeert Wekken vaak van het apparaat dat hem in slaap bracht. Wekken met Shift drukt een toets op het toetsenbord in, die meer hosts accepteren.',
      led: 'Power-LED',
      ledOn: 'Aan',
      ledOff: 'Uit',
      ledUnknown: 'Onbekend',
      ledConnected: 'Power-LED aangesloten',
      ledConnectedTip:
        'Schakel dit alleen in als de power-LED-aansluiting van de host met het bord is verbonden. Zonder die aansluiting is de stroomstatus onbekend.',
      ledConnectedFailed: 'Opslaan van de Power-LED-instelling mislukt',
      powerLongConfirm:
        'Aan/uit-knop {{seconds}} s ingedrukt houden? Dit schakelt de stroom uit zonder afsluiten.',
      done: 'Knop ingedrukt',
      failed: 'Knop indrukken mislukt'
    },
    settings: {
      title: 'Instellingen',
      mcp: {
        title: 'MCP-service',
        service: 'MCP-afstandsbediening',
        serviceDesc:
          'Vertrouwde MCP-clients toestaan het toetsenbord en de muis te bedienen en schermafbeeldingen te maken',
        securityWarning:
          'Iedereen met deze API-sleutel kan de externe host bedienen en het scherm bekijken. Gebruik HTTPS en schakel de service alleen in op vertrouwde netwerken.',
        endpoint: 'Eindpunt',
        apiKey: 'API-sleutel',
        regenerateConfirmTitle: 'MCP API-sleutel opnieuw genereren?',
        regenerateConfirmDesc: 'De huidige sleutel werkt dan onmiddellijk niet meer.',
        enableConfirmTitle: 'Externe MCP-bediening inschakelen?',
        enableConfirmDesc:
          'Als MCP wordt ingeschakeld, stopt PicoClaw en worden alle actieve PicoClaw-sessies gesloten.',
        failed: 'MCP-bewerking mislukt',
        copyFailed: 'Kopiëren mislukt. Kopieer handmatig.',
        okBtn: 'Bevestigen',
        cancelBtn: 'Annuleren'
      },
      redfish: {
        title: 'Redfish',
        service: 'Redfish-service',
        serviceDesc:
          'De DMTF Redfish API, voor stroombeheer, virtuele media en status vanuit tools zoals redfishtool en Ansible. Uitschakelen beëindigt alle Redfish-sessies.',
        endpoint: 'Service-root',
        httpsOn: 'Het bord gebruikt HTTPS, wat de meeste Redfish-tools nodig hebben.',
        httpsOff:
          'Het bord gebruikt onversleuteld HTTP. De meeste Redfish-tools hebben HTTPS nodig: schakel het in bij "Instellingen > Netwerk".',
        credentials:
          'Redfish accepteert de KVM-accounts, met Basic-authenticatie of een Redfish-sessie, en API-sleutels verzonden als X-Auth-Token. API-sleutels beheert u op de pagina API-sleutels.',
        powerActions: 'Stroomacties',
        powerActionsDesc:
          'De resettypen die nu worden aangeboden. On, ForceOff en GracefulShutdown hebben de stroomstatus nodig en worden daarom alleen aangeboden als "Power-LED aangesloten" is ingeschakeld in het stroommenu.',
        sessions: 'Sessies',
        noSessions: 'Geen open Redfish-sessies',
        created: 'Aangemaakt',
        lastUsed: 'Laatst gebruikt',
        refresh: 'Vernieuwen',
        end: 'Beëindigen',
        endConfirmTitle: 'Deze Redfish-sessie beëindigen?',
        endConfirmDesc: 'Het token werkt onmiddellijk niet meer. De client moet opnieuw inloggen.',
        failed: 'Redfish-actie mislukt',
        copyFailed: 'Kopiëren mislukt. Kopieer handmatig.',
        okBtn: 'Bevestigen',
        cancelBtn: 'Annuleren'
      },
      ipmi: {
        title: 'IPMI',
        warning:
          'IPMI-authenticatie is zwak door het ontwerp. Iedereen die het bord kan bereiken en een gebruikersnaam kent, kan een hash van het IPMI-wachtwoord van die gebruiker ophalen en offline proberen te kraken. Gebruik gegenereerde wachtwoorden, zet IPMI alleen aan op een vertrouwd netwerk en gebruik liever Redfish via HTTPS als een tool dat ondersteunt.',
        service: 'IPMI via LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) op UDP-poort 623, voor de voeding en status van de host. IPMI 1.5 en cipher suite 0 worden geweigerd. Uitzetten beëindigt alle IPMI-sessies.',
        example: 'Voorbeeld',
        copyFailed: 'Kopiëren mislukt. Kopieer handmatig.',
        ledOn: 'Voedingsstatus, on, off, soft, cycle en reset zijn beschikbaar.',
        ledOff:
          '"Power-LED aangesloten" staat uit in het voedingsmenu, dus de voedingstoestand is onbekend. Alleen "power reset" werkt: status, on, off, soft en cycle worden geweigerd.',
        accounts: 'Accounts',
        accountsDesc:
          'IPMI meldt zich aan met de KVM-accounts, elk met een eigen IPMI-wachtwoord, los van het webwachtwoord. Beheerders krijgen ADMINISTRATOR. Gebruikers krijgen USER: ze kunnen de voedingstoestand lezen met "-L USER", maar niet wijzigen.',
        passwordSet: 'IPMI-wachtwoord ingesteld',
        passwordNotSet: 'Geen IPMI-wachtwoord: kan niet aanmelden via IPMI',
        nameTooLong: 'De naam is langer dan 16 tekens, wat IPMI niet toestaat',
        accountDisabled: 'Het account is uitgeschakeld',
        setPassword: 'Wachtwoord instellen',
        changePassword: 'Wachtwoord wijzigen',
        remove: 'Verwijderen',
        removeConfirmTitle: 'Het IPMI-wachtwoord van {{user}} verwijderen?',
        removeConfirmDesc:
          'Het account kan zich niet meer via IPMI aanmelden en de IPMI-sessies ervan worden beëindigd.',
        passwordTitle: 'IPMI-wachtwoord voor {{user}}',
        passwordDesc:
          '12 tot 20 afdrukbare ASCII-tekens, anders dan het webwachtwoord. IPMI vereist dat het bord het wachtwoord bewaart in een vorm die het terug kan lezen, dus gebruik er een die nergens anders wordt gebruikt. Kopieer het voordat u opslaat: het wordt niet meer getoond.',
        passwordPlaceholder: 'IPMI-wachtwoord',
        generate: 'Genereren',
        copy: 'Kopiëren',
        save: 'Opslaan',
        passwordLength: 'Gebruik 12 tot 20 tekens.',
        passwordChars: 'Gebruik alleen afdrukbare ASCII-tekens.',
        saved: 'IPMI-wachtwoord opgeslagen',
        failed: 'IPMI-bewerking mislukt',
        okBtn: 'Bevestigen',
        cancelBtn: 'Annuleren'
      },
      vnc: {
        title: 'VNC',
        service: 'VNC-server',
        serviceDesc:
          'Laat een VNC-client, zoals TigerVNC of Remmina, de host bekijken en bedienen. De client moet Tight-codering ondersteunen. Eén sessie tegelijk.',
        credentials:
          'Meld u aan met een KVM-account. De verbinding is versleuteld met het TLS-certificaat van het bord (VeNCrypt X509Plain).',
        port: 'Poort',
        portDesc: 'De TCP-poort waarop de server luistert.',
        maxFps: 'Limiet beeldsnelheid',
        maxFpsDesc: 'Het maximale aantal beelden per seconde dat een client ontvangt.',
        vncAuth: 'Eenvoudige VNC-authenticatie',
        vncAuthDesc:
          'Voor clients zonder VeNCrypt. Deze controleert een apart VNC-wachtwoord in plaats van een account.',
        vncAuthWarning:
          'Eenvoudige VNC-authenticatie versleutelt de verbinding niet. Iedereen op het netwerkpad kan het scherm en de toetsaanslagen zien. Gebruik deze alleen op een vertrouwd netwerk.',
        password: 'VNC-wachtwoord',
        passwordSet: 'Er is een wachtwoord ingesteld. Typ een nieuw wachtwoord om het te wijzigen.',
        passwordInvalid: 'Het VNC-wachtwoord moet 6 tot 8 tekens lang zijn.',
        save: 'Opslaan',
        saved: 'Instellingen opgeslagen',
        state: 'Status',
        listening: 'Luistert op poort {{port}}',
        notListening: 'Luistert niet',
        noSession: 'Geen open sessie',
        client: 'Client',
        user: 'Gebruiker',
        method: 'Authenticatie',
        methodVencrypt: 'Account via TLS',
        methodVnc: 'VNC-wachtwoord',
        since: 'Verbonden sinds',
        resolution: 'Resolutie',
        framesSent: 'Verzonden beelden',
        lastError: 'De laatste sessie eindigde: {{error}}',
        refresh: 'Vernieuwen',
        disconnect: 'Verbreken',
        disconnectConfirmTitle: 'De VNC-sessie beëindigen?',
        disconnectConfirmDesc:
          'De client wordt meteen losgekoppeld, en elke ingedrukte toets en knop wordt losgelaten.',
        failed: 'VNC-bewerking mislukt',
        okBtn: 'Bevestigen',
        cancelBtn: 'Annuleren'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Host-watchdog',
        serviceDesc:
          'Als de host aan zou moeten staan en zijn beeld niet verandert, of er geen HDMI-signaal is, gedurende de time-out, drukt het bord op reset of zet het de host uit en weer aan.',
        stillWarning:
          'Een host waarvan het scherm in slaapstand gaat, of waarvan het beeld stilstaat terwijl hij werkt, lijkt vastgelopen. Zet de schermslaapstand op de host uit, of stel een ping-adres in.',
        ledHint:
          '"Power-LED aangesloten" staat uit in het stroommenu. De watchdog ziet niet wanneer de host uit staat, dus behandelt hij de host als altijd aan.',
        timeout: 'Time-out',
        timeoutDesc:
          'Hoe lang de host geen teken van leven mag geven voordat de watchdog ingrijpt.',
        action: 'Actie',
        actionDesc:
          'Uit en aan houdt de aan/uit-knop 5 seconden ingedrukt en drukt er daarna opnieuw op.',
        actionReset: 'Reset',
        actionPower: 'Uit en aan',
        cooldown: 'Wachttijd',
        cooldownDesc: 'De kortste tijd tussen twee acties.',
        maxPerHour: 'Acties per uur',
        maxPerHourDesc: 'Het maximale aantal acties in een uur.',
        pingHost: 'Ping-adres',
        pingHostDesc:
          'Het IP-adres van de host. Een antwoord telt als teken van leven. Laat leeg om niet te pingen.',
        pingHostInvalid: 'Voer een IPv4- of IPv6-adres in.',
        minutes: 'min',
        save: 'Opslaan',
        saved: 'Opgeslagen',
        state: 'Detector',
        status: {
          off: 'Uit',
          watching: 'Bewaakt',
          hostOff: 'Host uit',
          captureOff: 'HDMI-opname uit',
          cooldown: 'Wachttijd',
          capped: 'Uurlimiet bereikt',
          acting: 'Grijpt in'
        },
        signal: 'HDMI-signaal',
        yes: 'Ja',
        no: 'Nee',
        led: 'Power-LED',
        on: 'Aan',
        off: 'Uit',
        ledNotConnected: 'Niet aangesloten',
        ping: 'Ping',
        pingNotSet: 'Niet ingesteld',
        pingReply: 'Antwoordt',
        pingNoReply: 'Geen antwoord',
        lastChange: 'Laatste beeldwijziging',
        never: 'Nooit',
        actsIn: 'Grijpt in over',
        actionsLastHour: 'Acties in het afgelopen uur',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Logboek',
        noLog: 'De watchdog heeft nog niet ingegrepen.',
        refresh: 'Vernieuwen',
        reasonFrozen: 'Het beeld veranderde niet',
        reasonNoSignal: 'Geen HDMI-signaal',
        stuckFor: 'geen teken van leven gedurende {{duration}}',
        pressFailed: 'De druk op de knop mislukte: {{error}}',
        noScreenshot: 'Geen schermafbeelding',
        failed: 'Watchdog-bewerking mislukt',
        powerNeedsLed: 'Een stroomcyclus vereist "Aan/uit-LED aangesloten" in het aan/uit-menu.',
        noLedConfirmTitle: 'Watchdog inschakelen zonder aan/uit-LED?',
        noLedConfirmDesc:
          'Het bord ziet niet wanneer de host uit staat en behandelt de host daarom als altijd aan. Als u de host afsluit, drukt de watchdog na de time-out op reset. Sluit de aan/uit-LED aan om dit te voorkomen.',
        noLedConfirmOk: 'Inschakelen',
        cancel: 'Annuleren'
      },
      netboot: {
        title: 'Netwerkboot',
        description:
          'De host via het netwerk opstarten: iPXE en een menu van de images op de KVM via de USB-netwerkverbinding, of netboot.xyz via proxy-DHCP op het LAN.',
        addon: 'dnsmasq en opstartbestanden',
        addonDesc:
          'Geïnstalleerd op /data: dnsmasq uit Alpine, iPXE en netboot.xyz uit hun releases, elk gecontroleerd met de eigen checksum.',
        install: 'Installeren',
        installing: 'Bezig met installeren. Dit kan enkele minuten duren.',
        uninstall: 'Verwijderen',
        uninstallConfirm: 'Netwerkboot uitschakelen en dnsmasq en de opstartbestanden verwijderen?',
        needsData: 'Netwerkboot vereist een IronKVM-image met de /data-partitie aangekoppeld.',
        usb: 'Op de USB-netwerkverbinding',
        usbDesc:
          'Zolang de USB-netwerkverbinding aan staat, bedient dnsmasq die in plaats van udhcpd. De host krijgt zijn ene adres zonder router en zonder DNS-server, iPXE voor zijn architectuur en een menu van de ISO-images op de KVM.',
        linkOff: 'De USB-netwerkverbinding staat uit. Zet die aan onder Apparaat, USB-netwerk.',
        menuUrl: 'Menu',
        leases: 'Lease van de host',
        noLeases: 'Nog geen',
        netbootxyzNote:
          'netboot.xyz in het menu laadt van internet, dat de USB-verbinding niet bereikt. De host heeft daarvoor internet op een andere netwerkpoort nodig.',
        lan: 'Proxy-DHCP op het LAN',
        lanDesc:
          'Beantwoordt PXE-clients op het LAN met netboot.xyz, dat daarna zijn menu van internet laadt. Het deelt nooit adressen uit en biedt de images op de KVM niet aan.',
        lanWarning:
          'Elke PXE-client op dit LAN krijgt netboot.xyz aangeboden, niet alleen de host. Zet dit alleen aan op een netwerk dat je beheert.',
        lanConfirm: 'Proxy-DHCP op het LAN inschakelen?',
        lanInterface: 'LAN',
        running: 'Actief',
        stopped: 'Niet actief',
        images: 'Images in het menu',
        noImages: 'Geen ISO-images in de imagemap.',
        boots: 'Recente boots',
        noBoots: 'De host heeft nog niets opgehaald.',
        log: 'dnsmasq-logboek',
        refresh: 'Vernieuwen',
        okBtn: 'Bevestigen',
        cancelBtn: 'Annuleren',
        failed: 'Netwerkboot-bewerking mislukt'
      },
      about: {
        title: 'Over NanoKVM',
        information: 'Informatie',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Applicatie versie',
        applicationTip: 'Versie van de NanoKVM-webapplicatie',
        image: 'Image versie',
        imageTip: 'Versie van de NanoKVM-systeemimage',
        kernel: 'Kernelversie',
        kernelTip: 'Versie van de Linux-kernel die nu draait',
        deviceKey: 'Apparaat sleutel',
        videoMemory: 'Videogeheugen',
        videoMemoryTip:
          'Geheugen gereserveerd voor video-opname. Het wordt niet gedeeld met de rest van het systeem.',
        videoMemoryGenerations_one: '{{count}} eerdere NanoKVM-sessie houdt videogeheugen vast',
        videoMemoryGenerations_other: '{{count}} eerdere NanoKVM-sessies houden videogeheugen vast',
        videoMemoryReboot: 'Start opnieuw op om het vrij te maken.',
        community: 'Community',
        hostname: 'Hostnaam',
        hostnameUpdated: 'Hostnaam bijgewerkt. Start opnieuw op om toe te passen.',
        ipType: {
          Wired: 'Bedraad',
          Wireless: 'Draadloos',
          Other: 'Anders'
        },
        hostnameInvalid:
          'Gebruik letters, cijfers en koppeltekens, maximaal 63 per door punten gescheiden deel. Geen koppelteken aan het begin of einde van een deel.',
        hostnameFailed: 'De hostnaam kon niet worden gewijzigd'
      },
      appearance: {
        title: 'Uiterlijk',
        display: 'Beeldscherm',
        language: 'Taal',
        languageDesc: 'Selecteer de taal voor de interface',
        webTitle: 'Webtitel',
        webTitleDesc: 'Pas de titel van de webpagina aan',
        menuBar: {
          title: 'Menubalk',
          mode: 'Weergavemodus',
          modeDesc: 'Geef de menubalk weer op het scherm',
          modeOff: 'Uit',
          modeAuto: 'Automatisch verbergen',
          modeAlways: 'Altijd zichtbaar',
          keyboardLedStatus: 'Toetsvergrendelingsindicatoren',
          keyboardLedStatusDesc:
            'Toon de Num Lock-, Caps Lock- en Scroll Lock-status van de externe computer',
          icons: 'Submenupictogrammen',
          iconsDesc: 'Submenupictogrammen weergeven in de menubalk'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Toetsvergrendelingsstatus van extern toetsenbord',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Aan',
        off: 'Uit',
        unknown: 'Onbekend'
      },
      device: {
        title: 'Apparaat',
        oled: {
          title: 'OLED',
          description: 'OLED scherm automatisch slapen',
          brightness: 'OLED-helderheid',
          brightnessDescription: 'Een lager niveau verlengt de levensduur van het scherm',
          brightnessLevels: {
            '64': 'Laagst',
            '96': 'Laag',
            '128': 'Gemiddeld',
            '160': 'Hoog',
            '207': 'Standaard',
            '255': 'Maximaal'
          },
          0: 'Nooit',
          15: '15 sec',
          30: '30 sec',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 uur'
        },
        ssh: {
          description: 'Schakel SSH externe toegang in',
          tip: 'Stel een sterk wachtwoord in voordat u (Account - Wachtwoord wijzigen) inschakelt'
        },
        advanced: 'Geavanceerde instellingen',
        cpuFreq: {
          title: 'CPU-frequentie',
          description: 'Stel de CPU-kloksnelheid in voor de volgende keer opstarten',
          tip: 'De CPU start op 850 MHz en is gespecificeerd voor 1000 MHz. Een nieuwe waarde wordt bij de volgende keer opstarten toegepast, niet terwijl het systeem draait. 1000 MHz valt binnen de specificatie; de temperatuur blijft bij beide instellingen ruim binnen de grenzen.',
          running: 'Actief: {{mhz}} MHz',
          rebootToApply: 'opnieuw opstarten om toe te passen',
          rebootConfirm: 'Nu opnieuw opstarten om {{mhz}} MHz toe te passen?'
        },
        swap: {
          title: 'Wisselen',
          disable: 'Uitschakelen',
          description: 'Stel de grootte van het wisselbestand in',
          tip: 'Het inschakelen van deze functie kan de bruikbare levensduur van uw SD-kaart verkorten!'
        },
        zram: {
          title: 'Gecomprimeerde swap (zram)',
          description: 'Swap in gecomprimeerd RAM in plaats van op de SD-kaart',
          tip: 'zram houdt swap van de SD-kaart af en veroorzaakt dus geen slijtage. Er zit geen swap op schijf achter: als zram vol raakt, stopt de kernel een proces in plaats van traag te pagineren. De geheugenlimiet bepaalt hoeveel RAM zram mag gebruiken.',
          unavailable: 'De kernelmodules zijn niet geïnstalleerd op dit apparaat',
          inactive: 'Ingeschakeld, maar het apparaat is niet gestart',
          active: 'Actief - {{used}} van {{total}}, {{ratio}}x',
          off: 'Uit',
          detail: {
            algorithm: 'Algoritme: {{algorithm}}',
            memory: 'Geheugen gebruikt: {{used}} van {{limit}}',
            memoryNoLimit: 'Geheugen gebruikt: {{used}}, geen limiet ingesteld',
            counters:
              "Pagina's ingeswapt {{in}}, uitgeswapt {{out}} (alle swap-apparaten, sinds opstarten)"
          }
        },
        mouseJiggler: {
          title: 'Muisschommel',
          description: 'Voorkom dat de externe host in slaap valt',
          disable: 'Uitschakelen',
          absolute: 'Absolute modus',
          relative: 'Relatieve modus'
        },
        mdns: {
          description: 'Schakel de mDNS detectieservice in',
          tip: 'Schakel het uit als het niet nodig is'
        },
        hdmi: {
          description: 'Schakel HDMI/monitoruitgang in',
          idleTimeoutTitle: 'Time-out voor inactieve opname',
          idleTimeoutDescription:
            'HDMI-opname stoppen nadat er gedurende deze tijd geen actieve kijkers zijn:',
          minutes: 'min'
        },
        autostart: {
          title: 'Instellingen voor automatisch starten van scripts',
          description:
            'Beheer scripts die automatisch worden uitgevoerd bij het opstarten van het systeem',
          new: 'Nieuw',
          deleteConfirm: 'Weet u zeker dat u dit bestand wilt verwijderen?',
          yes: 'Ja',
          no: 'Nee',
          scriptName: 'Scriptnaam automatisch starten',
          scriptContent: 'Scriptinhoud automatisch starten',
          settings: 'Instellingen'
        },
        hidOnly: 'HID-Alleen modus',
        hidOnlyDesc:
          'Stop met het emuleren van virtuele apparaten en behoud alleen de basisbesturing van HID',
        disk: 'Virtuele schijf',
        diskDesc: 'Koppel virtuele U-schijf aan de externe host',
        network: 'Virtueel Netwerk',
        networkDesc: 'Koppel virtueel netwerk kaart aan de externe host',
        usbNetwork: {
          description:
            'Een privé netwerkverbinding met de externe host via de USB-kabel. De host krijgt een adres zonder gateway en zonder DNS, en kan uw LAN dus niet via NanoKVM bereiken.',
          off: 'Uit',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (voor hosts zonder NCM)',
          rndis: 'RNDIS (niet meer aangeboden)',
          rndisNote:
            'Deze verbinding gebruikt RNDIS, dat niet meer wordt aangeboden. Kies NCM of ECM.',
          subnet: 'Subnet',
          subnetDesc:
            'Een privé IPv4-netwerk, /24 tot /30. NanoKVM neemt het eerste adres, de host het tweede.',
          addresses: 'NanoKVM: {{board}}, host: {{host}}',
          invalidSubnet: 'Voer een subnet in, zoals 172.31.255.0/30.',
          apply: 'Toepassen',
          confirm: 'USB-apparaat opnieuw verbinden?',
          reenumerate:
            'Toepassen bouwt de USB-verbinding opnieuw op. De host verliest het toetsenbord, de muis en de virtuele schijf enkele seconden.'
        },
        audio: 'Virtuele luidspreker',
        audioDesc:
          'Biedt de externe host een USB-geluidskaart aan, zodat u het geluid kunt horen. De host moet deze als uitvoerapparaat kiezen. Omschakelen bouwt de USB-verbinding opnieuw op.',
        audioNote: 'Audio is beschikbaar in beide H.264-modi (WebRTC en Direct), niet in MJPEG',
        console: 'Seriële console',
        consoleDesc:
          'Biedt de externe host een seriële USB-poort aan, om op deze NanoKVM in te loggen als het netwerk onbereikbaar is',
        consoleTip:
          'Iedereen die de externe host bedient, krijgt een inlogprompt van deze NanoKVM. Stel een sterk wachtwoord in voordat u dit inschakelt (Account - Wachtwoord wijzigen).',
        endpoints: {
          title: 'USB-endpoints',
          used: '{{used}} van {{total}} gebruikt',
          cost: 'gebruikt {{cost}}',
          needs: 'heeft {{cost}} nodig',
          full: 'Niet genoeg USB-endpoints. Schakel eerst iets anders uit.',
          inactive:
            'Aan, maar niet actief: de USB-controller heeft geen endpoints meer. Schakel een ander apparaat uit en dit apparaat start meteen.',
          explain:
            'De USB-controller heeft een vast aantal inkomende endpoints, en die worden hier geteld. Als er meer apparaten zijn ingeschakeld dan er passen, blijven toetsenbord en muis behouden en wordt de rest uitgeschakeld.',
          error: 'Kan het apparaat niet bereiken. Probeer het opnieuw.',
          fitTogether: 'Deze passen samen: {{sets}}'
        },
        reboot: 'Opnieuw opstarten',
        rebootDesc: 'Weet u zeker dat u NanoKVM opnieuw wilt opstarten?',
        okBtn: 'Ja',
        cancelBtn: 'Nee',
        rebootFailed: 'Herstarten mislukt'
      },
      network: {
        title: 'Netwerk',
        wifi: {
          title: 'Wi-Fi',
          description: 'Wi-Fi configureren',
          apMode: 'AP-modus is ingeschakeld, maak verbinding met Wi-Fi door de QR-code te scannen',
          connect: 'Wi-Fi verbinden',
          connectDesc1: 'Voer de netwerk-SSID en het wachtwoord in',
          connectDesc2: 'Voer het wachtwoord in om met dit netwerk te verbinden',
          disconnect: 'Weet je zeker dat je de netwerkverbinding wilt verbreken?',
          failed: 'Verbinding mislukt, probeer het opnieuw.',
          ssid: 'Naam',
          password: 'Wachtwoord',
          joinBtn: 'Verbinden',
          confirmBtn: 'OK',
          cancelBtn: 'Annuleren'
        },
        tls: {
          description: 'HTTPS-protocol inschakelen',
          tip: 'Let op: HTTPS gebruiken kan de latentie verhogen, vooral in MJPEG-videomodus.',
          restarting: 'De apparaatserver wordt opnieuw gestart, dit duurt ongeveer twee minuten...',
          waiting: 'Wachten tot het apparaat weer reageert...',
          waitingHttp: 'Terugschakelen naar http. Vernieuw deze pagina als deze niet vanzelf opent.',
          failed: 'De HTTPS-instelling kon niet worden gewijzigd',
          enableConfirm: 'HTTPS inschakelen?',
          disableConfirm: 'HTTPS uitschakelen?',
          confirmDesc:
            'Dit meldt u af en herstart de server van het apparaat, wat ongeveer twee minuten duurt. Daarna opent de pagina {{url}}.',
          confirmOk: 'Doorgaan',
          confirmCancel: 'Annuleren'
        },
        ethernet: {
          title: 'IP-adres',
          description: 'Stel in hoe NanoKVM zijn adres op het bekabelde netwerk krijgt',
          dhcp: 'DHCP',
          manual: 'Handmatig',
          networkDetails: 'Netwerkgegevens',
          interface: 'Interface',
          ipAddress: 'IP-adres',
          subnetMask: 'Subnetmasker',
          router: 'Router',
          save: 'Toepassen',
          invalidAddress: 'Voer een geldig IP-adres in',
          invalidMask: 'Voer een geldig subnetmasker in, bijvoorbeeld 255.255.255.0 of 24',
          invalidRouter: 'Voer een geldig routeradres in',
          addressRequired: 'Een IP-adres is vereist',
          maskRequired: 'Een subnetmasker is vereist',
          applyTitle: 'Het adres van NanoKVM wijzigen?',
          applyWarning:
            'De verbinding met deze pagina gaat verloren. NanoKVM past het nieuwe adres toe en wacht {{seconds}} seconden tot u het daar bereikt. Bereiken behoudt de wijziging. Bereikt niets het, dan zet NanoKVM de vorige instellingen terug.',
          applyConfirm: 'Toepassen',
          applyCancel: 'Annuleren',
          applyFailed: 'Het adres kon niet worden toegepast',
          trialTitle: 'Wacht op bevestiging',
          trialDhcp: 'NanoKVM vraagt een adres aan via DHCP.',
          trialStatic: 'NanoKVM is nu bereikbaar op {{address}}.',
          trialInstruction:
            'Open NanoKVM op zijn nieuwe adres en meld u aan als daarom wordt gevraagd. Daar bereiken behoudt de wijziging. Bereikt niets NanoKVM binnen {{seconds}} seconden, dan zet het de vorige instellingen terug.',
          trialOpen: 'Het nieuwe adres openen',
          trialKeep: 'Deze instellingen behouden',
          trialKept: 'Het nieuwe adres is opgeslagen',
          trialKeepFailed: 'De instellingen konden niet worden behouden',
          trialGone: 'De wijziging is al teruggezet. Probeer het opnieuw.',
          unsaved: 'Niet-opgeslagen wijzigingen'
        },
        dns: {
          title: 'DNS',
          description: 'Configureer DNS-servers voor NanoKVM',
          mode: 'Modus',
          dhcp: 'DHCP',
          manual: 'Handmatig',
          add: 'DNS toevoegen',
          save: 'Opslaan',
          invalid: 'Voer een geldig IP-adres in',
          noDhcp: 'Er is momenteel geen DHCP-DNS beschikbaar',
          saved: 'DNS-instellingen opgeslagen',
          saveFailed: 'DNS-instellingen opslaan mislukt',
          unsaved: 'Niet-opgeslagen wijzigingen',
          maxServers: 'Maximaal {{count}} DNS-servers toegestaan',
          dnsServers: 'DNS-servers',
          dhcpServersDescription: 'DNS-servers worden automatisch via DHCP verkregen',
          manualServersDescription: 'DNS-servers kunnen handmatig worden bewerkt',
          networkDetails: 'Netwerkdetails',
          interface: 'Interface',
          ipAddress: 'IP-adres',
          subnetMask: 'Subnetmasker',
          router: 'Router',
          none: 'Geen'
        }
      },
      vpn: {
        loading: 'Laden...',
        okBtn: 'Ja',
        cancelBtn: 'Nee',
        restart: '{{name}} opnieuw starten?',
        stop: '{{name}} stoppen?',
        stopDesc:
          'De daemon stopt nu. "Starten bij opstarten" is een aparte schakelaar en blijft zoals hij is.',
        update: '{{name}} bijwerken naar {{version}}?',
        updateDesc: 'De daemon herstart als deze draait. De aanmelding blijft behouden.',
        notInstall: '{{name}} is niet geïnstalleerd.',
        install: 'Installeren',
        installing: 'Installeren',
        installFailed: 'Installatie mislukt',
        retry: 'Opnieuw proberen',
        notRunning: '{{name}} draait niet. Start het om verder te gaan.',
        run: 'Starten',
        boot: 'Starten bij opstarten',
        bootDesc: '{{name}} starten wanneer de KVM opstart.',
        enable: '{{name}} inschakelen',
        control: 'Controleserver',
        connected: 'Verbonden',
        disconnected: 'Niet verbonden',
        deviceName: 'Apparaatnaam',
        deviceIP: 'Apparaat-IP',
        account: 'Account',
        version: 'Versie',
        uptime: 'Uptime',
        peers: 'Peers',
        noPeers: 'Nog geen peers.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Geheugen',
        daemonRss: 'Daemon',
        group: 'Add-ongroep',
        high: 'afgeremd boven {{size}}',
        max: 'gestopt door de kernel boven {{size}}',
        noGroup: 'Geen geheugengroep voor add-ons op dit bord.',
        uninstall: '{{name}} verwijderen',
        uninstallDesc:
          'Weet u zeker dat u {{name}} wilt verwijderen? De aanmelding blijft op het bord bewaard.',
        blocked:
          '{{other}} draait of start bij het opstarten. Er kan maar één VPN tegelijk draaien: stop eerst {{other}} en schakel het starten bij opstarten ervan uit.',
        swap: {
          title: 'Swapgeheugen',
          tip: 'Als de daemon te weinig geheugen heeft, probeer dan swapgeheugen in te schakelen. Dit stelt het wisselbestand standaard in op 256MB; de grootte kunt u aanpassen in "Instellingen > Apparaat".'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Vernieuw en probeer opnieuw. Of probeer handmatig te installeren',
        download: 'Download het',
        package: 'installatiepakket',
        unzip: 'en pak het uit',
        upTailscale: 'Upload tailscale naar NanoKVM directory /usr/bin/',
        upTailscaled: 'Upload tailscaled naar NanoKVM directory /usr/sbin/',
        refresh: 'Vernieuw huidige pagina',
        notLogin:
          'Het apparaat is nog niet gekoppeld. Log in en koppel dit apparaat aan uw account.',
        urlPeriod: 'Deze url is 10 minuten geldig',
        login: 'Inloggen',
        loginSuccess: 'Inloggen gelukt',
        logout: 'Uitloggen',
        logoutDesc: 'Weet u zeker dat u wilt uitloggen?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Dit apparaat is nog niet lid van een NetBird-netwerk. Word lid met een setup key of log in met SSO.',
        setupKey: 'Setup key',
        setupKeyPlaceholder: 'Plak een setup key uit het NetBird-dashboard',
        join: 'Lid worden',
        or: 'of',
        sso: 'Inloggen met SSO',
        urlPeriod: 'Deze url is 10 minuten geldig',
        loginSuccess: 'Inloggen gelukt',
        logout: 'Afmelden',
        logoutDesc:
          'Afmelden verwijdert deze peer uit uw NetBird-account en wist de configuratie hier. Opnieuw lid worden vereist een setup key of een SSO-login, en de peer kan een nieuw IP-adres krijgen. Doorgaan?'
      },
      update: {
        title: 'Controleren op updates',
        queryFailed: 'Ophalen versie mislukt',
        updateFailed: 'Update mislukt. Probeer het opnieuw.',
        isLatest: 'U heeft al de nieuwste versie.',
        available: 'Er is een update beschikbaar. Weet u zeker dat u wilt updaten?',
        updating: 'Update gestart. Even geduld a.u.b...',
        confirm: 'Bevestigen',
        cancel: 'Annuleren',
        preview: 'Preview-updates',
        previewDesc: 'Krijg vroegtijdig toegang tot nieuwe functies en verbeteringen',
        previewTip:
          'Houd er rekening mee dat preview-releases bugs of onvolledige functionaliteit kunnen bevatten!',
        customServer: {
          title: 'Aangepaste updateserver',
          desc: 'Online-updates zoeken en downloaden vanaf een opgegeven server',
          invalidUrl:
            'Voer een geldige HTTP- of HTTPS-servermap in zonder queryparameters, fragment of latest.json.',
          loadFailed: 'De configuratie van de updateserver kon niet worden geladen.',
          saveFailed: 'De configuratie van de updateserver kon niet worden opgeslagen.',
          saved: 'De configuratie van de updateserver is opgeslagen.',
          save: 'Opslaan',
          confirmTitle: 'Een aangepaste updateserver gebruiken?',
          confirmDesc:
            'SHA-512 controleert alleen of het pakket overeenkomt met het manifest dat door deze server wordt verstrekt. Het bewijst niet dat het pakket een officiële NanoKVM-release is. Een defecte of kwaadwillende server kan het apparaat onbruikbaar maken, gegevensverlies veroorzaken of het systeem compromitteren.',
          confirm: 'Toch gebruiken',
          useSipeed: 'De officiële Sipeed-server gebruiken',
          previewDisabled:
            'Preview-updates zijn niet beschikbaar zolang een aangepaste updateserver is ingeschakeld.'
        },
        offline: {
          title: 'Offline-updates',
          desc: 'Update via lokaal installatiepakket',
          upload: 'Uploaden',
          checksumPlaceholder: 'SHA-256-controlesom (optioneel)',
          invalidChecksum: 'De SHA-256-controlesom moet 64 hexadecimale tekens bevatten.',
          checksumMismatch: 'De SHA-256-verificatie is mislukt. Het pakket is mogelijk beschadigd.',
          invalidName: 'Ongeldig bestandsnaamformaat. Download de versie van GitHub-releases.',
          updateFailed: 'Update mislukt. Probeer het opnieuw.'
        },
        updateTo: 'Bijwerken naar {{version}}',
        updateConfirmDesc:
          'Het apparaat installeert de update en herstart zijn server. Deze pagina herlaadt zodra de server terug is.'
      },
      account: {
        title: 'Account',
        webAccount: 'Web Account Naam',
        role: 'Rol',
        roles: { admin: 'Beheerder', user: 'Gebruiker' },
        password: 'Wachtwoord',
        updateBtn: 'Update',
        logoutBtn: 'Afmelden',
        logoutDesc: 'Weet u zeker dat u wilt uitloggen?',
        okBtn: 'Ja',
        cancelBtn: 'Nee',
        users: {
          title: 'Gebruikers',
          create: 'Gebruiker aanmaken',
          enabled: 'Ingeschakeld',
          disabled: 'Uitgeschakeld',
          deviceOwner: 'Eigenaar van het apparaat',
          resetPassword: 'Wachtwoord resetten',
          delete: 'Verwijderen',
          deleteConfirm: 'Deze gebruiker verwijderen en al zijn sessies intrekken?',
          created: 'Gebruiker aangemaakt',
          deleted: 'Gebruiker verwijderd',
          passwordUpdated: 'Wachtwoord bijgewerkt',
          loadFailed: 'Laden van gebruikers mislukt',
          saveFailed: 'Opslaan van gebruiker mislukt',
          deleteFailed: 'Verwijderen van gebruiker mislukt'
        }
      },
      apiKeys: {
        title: 'API-sleutels',
        description:
          'Een sleutel handelt namens zijn eigenaar, met de rol van die gebruiker. Stuur hem als Authorization: Bearer <key> voor metrics en de API, of als X-Auth-Token voor Redfish.',
        name: 'Naam',
        namePlaceholder: 'Waarvoor de sleutel is, bijvoorbeeld prometheus',
        nameRequired: 'Geef de sleutel een naam',
        nameTooLong: 'De naam mag maximaal 64 tekens lang zijn',
        unnamed: '(naamloos)',
        create: 'Sleutel aanmaken',
        created: 'Aangemaakt',
        owner: 'Eigenaar',
        empty: 'Geen API-sleutels',
        newKeyTitle: 'Uw nieuwe API-sleutel',
        newKeyWarning:
          'Kopieer de sleutel nu. Hij wordt niet opgeslagen en kan niet opnieuw worden getoond. Als u hem kwijtraakt, trek hem dan in en maak een nieuwe aan.',
        copy: 'Kopiëren',
        copied: 'Gekopieerd',
        copyFailed: 'Kopiëren mislukt. Kopieer handmatig.',
        done: 'Klaar',
        revoke: 'Intrekken',
        revokeConfirmTitle: 'Deze API-sleutel intrekken?',
        revokeConfirmDesc: 'Alles wat "{{name}}" gebruikt, werkt onmiddellijk niet meer.',
        revoked: 'API-sleutel ingetrokken',
        loadFailed: 'Laden van API-sleutels mislukt',
        createFailed: 'Aanmaken van API-sleutel mislukt',
        revokeFailed: 'Intrekken van API-sleutel mislukt',
        cancelBtn: 'Annuleren'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistent',
      empty: 'Open het paneel en start een taak om te beginnen.',
      inputPlaceholder: 'Beschrijf wat u wilt dat de PicoClaw doet',
      newConversation: 'Nieuw gesprek',
      processing: 'Verwerken...',
      agent: {
        defaultTitle: 'Algemene assistent',
        defaultDescription: 'Algemene hulp bij chatten, zoeken en werkruimte.',
        kvmTitle: 'Bediening op afstand',
        kvmDescription: 'Bedien de externe host via NanoKVM.',
        switched: 'Agentrol gewijzigd',
        switchFailed: 'Kan agentrol niet wisselen'
      },
      send: 'Verzenden',
      cancel: 'Annuleren',
      status: {
        connecting: 'Verbinden met gateway...',
        connected: 'PicoClaw-sessie verbonden',
        disconnected: 'PicoClaw-sessie gesloten',
        stopped: 'Stopverzoek verzonden',
        runtimeStarted: 'PicoClaw runtime gestart',
        runtimeStartFailed: 'Kan PicoClaw runtime niet starten',
        runtimeStopped: 'PicoClaw runtime gestopt',
        runtimeStopFailed: 'Kan PicoClaw runtime niet stoppen',
        controlSwitchedToMCP: 'Bediening overgeschakeld naar de externe MCP-service'
      },
      connection: {
        runtime: {
          checking: 'Controleren',
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime gereed',
          stopped: 'Runtime gestopt',
          blockedByMCP: 'Externe MCP-bediening is actief',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime niet beschikbaar',
          configError: 'Configuratiefout'
        },
        transport: {
          connecting: 'Verbinden',
          connected: 'Verbonden',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'Inactief',
          busy: 'Bezet'
        }
      },
      message: {
        toolAction: 'Actie',
        observation: 'Observatie',
        screenshot: 'Schermafbeelding'
      },
      overlay: {
        locked: 'PicoClaw bestuurt het apparaat. Handmatige invoer is gepauzeerd.'
      },
      control: {
        picoclaw: 'Apparaatbediening: PicoClaw',
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Apparaatbediening: externe MCP',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Apparaatbediening: uit',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Bediening geven',
        release: 'Vrijgeven',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'PicoClaw-bediening gegeven',
        released: 'PicoClaw-bediening vrijgegeven',
        grantFailed: 'Kan PicoClaw-bediening niet geven',
        releaseFailed: 'Kan PicoClaw-bediening niet vrijgeven',
        grantConfirmTitle: 'Apparaatbediening overschakelen naar PicoClaw?',
        grantConfirmDesc: 'Schrijfacties van de externe MCP naar het apparaat worden onderbroken.'
      },
      install: {
        install: 'PicoClaw installeren',
        installing: 'PicoClaw installeren',
        success: 'PicoClaw is succesvol geïnstalleerd',
        failed: 'Kan PicoClaw niet installeren',
        uninstalling: 'Runtime verwijderen...',
        uninstalled: 'Runtime is succesvol verwijderd.',
        uninstallFailed: 'Verwijderen mislukt.',
        requiredTitle: 'PicoClaw is niet geïnstalleerd',
        requiredDescription: 'Installeer PicoClaw voordat u de runtime van PicoClaw start.',
        progressDescription: 'PicoClaw wordt gedownload en geïnstalleerd.',
        stages: {
          preparing: 'Voorbereiden',
          downloading: 'Downloaden',
          extracting: 'Uitpakken',
          verifying: 'Verifiëren',
          installing: 'Installeren',
          installed: 'Geïnstalleerd',
          install_timeout: 'Time-out',
          install_failed: 'Mislukt'
        }
      },
      model: {
        requiredTitle: 'Modelconfiguratie is vereist',
        requiredDescription: 'Configureer het PicoClaw-model voordat u PicoClaw chat gebruikt.',
        docsTitle: 'Configuratiehandleiding',
        docsDesc: 'Ondersteunde modellen en protocollen',
        menuLabel: 'Model configureren',
        modelIdentifier: 'Modelidentificatie',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API-sleutel',
        apiKeyPlaceholder: 'Voer de API-sleutel van het model in',
        save: 'Opslaan',
        saving: 'Opslaan',
        saved: 'Modelconfiguratie opgeslagen',
        saveFailed: 'Kan de modelconfiguratie niet opslaan',
        invalid: 'Model-ID, API Base URL en API-sleutel zijn vereist'
      },
      uninstall: {
        menuLabel: 'Verwijderen',
        confirmTitle: 'PicoClaw verwijderen',
        confirmContent:
          'Weet u zeker dat u PicoClaw wilt verwijderen? Hiermee worden het uitvoerbare bestand en alle configuratiebestanden verwijderd.',
        confirmOk: 'Verwijderen',
        confirmCancel: 'Annuleren'
      },
      history: {
        title: 'Geschiedenis',
        loading: 'Sessies laden...',
        emptyTitle: 'Nog geen geschiedenis',
        emptyDescription: 'Eerdere PicoClaw sessies verschijnen hier.',
        loadFailed: 'Kan de sessiegeschiedenis niet laden',
        deleteFailed: 'Kan sessie niet verwijderen',
        deleteConfirmTitle: 'Sessie verwijderen',
        deleteConfirmContent: 'Weet u zeker dat u "{{title}}" wilt verwijderen?',
        deleteConfirmOk: 'Verwijderen',
        deleteConfirmCancel: 'Annuleren',
        messageCount_one: '{{count}} bericht',
        messageCount_other: '{{count}} berichten',
        messageCount: '{{count}} berichten'
      },
      config: {
        startRuntime: 'Start PicoClaw',
        stopRuntime: 'Stop PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Bediening overschakelen naar PicoClaw?',
        enableConfirmDesc:
          'Bij het starten van PicoClaw wordt de externe MCP-service uitgeschakeld.',
        enableConfirmOk: 'PicoClaw starten',
        enableConfirmCancel: 'Annuleren',
        title: 'Start PicoClaw',
        description: 'Start de runtime om de PicoClaw assistent te gaan gebruiken.',
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'Er is een probleem opgetreden',
      refresh: 'Vernieuwen',
      panel: 'Dit deel van de pagina werkt niet meer',
      retry: 'Opnieuw proberen'
    },
    fullscreen: {
      toggle: 'Volledig scherm schakelen'
    },
    input: {
      disconnected: 'Toetsenbord en muis zijn niet verbonden',
      disconnectedTls:
        'De browser weigerde de beveiligde verbinding voor toetsenbord en muis, en doet dat zonder te vragen. Het certificaat dat dit apparaat heeft gegenereerd, wordt nog niet vertrouwd. Open dit adres in een nieuw tabblad, accepteer het certificaat en vernieuw daarna de pagina. Het certificaat installeren is de betrouwbare oplossing.',
      disconnectedNever:
        'De verbinding voor toetsenbord en muis kon niet worden geopend. De rest van de pagina werkt, omdat die deze verbinding niet gebruikt. Controleer of niets tussen u en het apparaat haar blokkeert.',
      disconnectedDropped:
        'De verbinding voor toetsenbord en muis is verbroken en niet hersteld. Na een herstart maakt ze vanzelf opnieuw verbinding; als dit blijft, vernieuw dan de pagina.',
      hidDisabled: 'HID is uitgeschakeld op dit apparaat (/boot/disable_hid).',
      keyFailed: 'De toets kon niet worden verzonden.'
    },
    speaker: { title: 'Luidspreker', unmute: 'Geluid aan', mute: 'Dempen' },
    menu: {
      collapse: 'Menu samenvouwen',
      expand: 'Menu uitvouwen'
    },
    ion: {
      checking: 'Videogeheugen controleren voordat de stream start...',
      warn: 'Het videogeheugen raakt op. Eén herstart van de server zou het uitputten. Start opnieuw op wanneer het u uitkomt.',
      criticalTitle: 'Niet genoeg videogeheugen om de stream te starten',
      criticalBody:
        'Video starten zou het gereserveerde geheugen uitputten en de server stoppen. Alle andere functies blijven werken, inclusief stroombeheer en opnieuw opstarten. Alleen opnieuw opstarten van de NanoKVM maakt dit geheugen vrij.',
      criticalContinue: 'Video toch starten',
      criticalReboot: 'NanoKVM opnieuw opstarten',
      criticalRebooting: 'Opnieuw opstarten...'
    }
  }
};

export default nl;
