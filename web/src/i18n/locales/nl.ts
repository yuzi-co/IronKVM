const nl = {
  translation: {
    feedback: {
      enabled: '{{name}} ingeschakeld',
      disabled: '{{name}} uitgeschakeld',
      failed: 'Het verzoek is mislukt. Probeer het opnieuw.',
      network: 'Het apparaat is niet bereikbaar. Controleer de verbinding en probeer het opnieuw.',
      saved: 'Opgeslagen',
      timeout: 'Het apparaat deed er te lang over om te antwoorden. Probeer het opnieuw.'
    },
    common: {
      copy: 'Kopiëren',
      copied: 'Gekopieerd',
      copyFailed: 'Kopiëren mislukt. Selecteer de tekst en kopieer hem handmatig.',
      notUpdating: 'Wordt niet bijgewerkt: de laatste verversing is mislukt.',
      off: 'Uit',
      running: 'Actief',
      save: 'Opslaan',
      cancel: 'Annuleren',
      delete: 'Verwijderen',
      remove: 'Weghalen'
    },
    head: {
      desktop: 'Extern bureaublad',
      login: 'Inloggen',
      changePassword: 'Wachtwoord wijzigen',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Wachtwoord gewijzigd. Meld je aan met het nieuwe wachtwoord.',
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
          'Om de wachtwoorden opnieuw in te stellen, houdt u de BOOT-knop op de IronKVM 10 seconden lang ingedrukt.',
        reset3: 'Standaard webaccount:',
        reset4: 'Standaard SSH-account:',
        change1: 'Houd er rekening mee dat deze actie de volgende wachtwoorden zal wijzigen:',
        change2: 'Web login wachtwoord',
        change3: 'Systeem root-wachtwoord (SSH-inlogwachtwoord)',
        change4:
          'Om de wachtwoorden opnieuw in te stellen, houdt u de BOOT-knop op de IronKVM ingedrukt.',
        resetDocs: 'Zie de hardwaredocumentatie voor de stappen in detail:',
        hardwareDocs: 'Sipeed NanoKVM-wiki'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Wi-Fi configureren voor IronKVM',
      success: 'Controleer de netwerkstatus van IronKVM en bezoek het nieuwe IP-adres.',
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
      },
      ssidRequired: 'Voer de netwerknaam in, maximaal 32 tekens',
      passwordLength: 'Het wachtwoord is 8 tot 63 tekens. Laat het leeg voor een open netwerk.',
      passwordOptional: 'Wachtwoord (leeg voor een open netwerk)',
      lost: 'Het bord reageert niet meer. Het heeft zich misschien bij het netwerk aangesloten en zijn instellingshotspot gesloten. Komt de hotspot terug, dan is verbinden mislukt: maak opnieuw verbinding en probeer het nog eens.',
      done: 'Instellen voltooid. Verbind dit apparaat weer met uw gebruikelijke netwerk en open het bord op zijn nieuwe adres.'
    },
    screen: {
      viewOnly: 'Alleen kijken',
      viewOnlyTip:
        'Dit tabblad stuurt geen toetsenbord- en muisinvoer meer naar de host. Scripts, de muis-jiggler en andere kijkers merken er niets van.',
      viewOnlyOff: 'Alleen kijken uitzetten',
      viewOnlyBlocked: 'Alleen kijken staat aan, er is niets naar de host gestuurd',
      pauseHidden: 'Pauzeren als tabblad verborgen is',
      pauseHiddenTip:
        'Stopt beeld en geluid enkele seconden nadat dit tabblad verborgen is en start ze weer als u terugkomt.',
      screenshot: 'Schermafbeelding',
      screenshotTip: 'Slaat het scherm van de host op als PNG op volle opnamegrootte.',
      screenshotFailed: 'Schermafbeelding mislukt',
      stream: {
        ok: 'beeld OK',
        noSignal: 'geen signaal',
        failed: 'stream mislukt'
      },
      codecNoWebrtcHevc: 'Deze browser kan H.265 niet via WebRTC ontvangen',
      codecNoHevc: 'Deze browser kan H.265 niet decoderen',
      codecNote:
        'Het bord heeft één encoder, dus dit wijzigt de stream voor alle kijkers. Maak opnieuw verbinding om het op een lopende WebRTC-sessie toe te passen.',
      codec: 'Codec',
      updateFailed: 'De instelling is niet toegepast',
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
      qualityLossless: 'Beste',
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
      },
      directConnectionFailed: 'Verbinding met de videostream mislukt'
    },
    keyboard: {
      close: 'Sluiten',
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
        sendFailed: 'Niet verzonden: de invoerverbinding is verbroken',
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
        saveFailed: 'Leadertoets kon niet worden opgeslagen',
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
      jiggler: 'Muisbeweger',
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
      scrollUp: 'Zoals op deze computer',
      scrollDown: 'Omgekeerd (natuurlijk scrollen)',
      speed: 'Scrollwielsnelheid',
      fast: 'Snel',
      slow: 'Langzaam',
      requestPointer:
        'Relatieve modus wordt gebruikt. Klik op het bureaublad om de muisaanwijzer te krijgen.',
      resetHid: 'HID resetten',
      hidOnly: {
        switchFailed: 'Modus wisselen mislukt. Controleer de verbinding en probeer het opnieuw.',
        title: 'Alleen HID-modus',
        desc: 'Als uw muis en toetsenbord niet meer reageren en het opnieuw instellen van HID niet helpt, kan er sprake zijn van een compatibiliteitsprobleem tussen de IronKVM en het apparaat. Probeer de modus HID-Only in te schakelen voor betere compatibiliteit.',
        tip1: 'Als u de modus HID-Only inschakelt, worden de virtuele U-schijf en het virtuele netwerk ontkoppeld',
        tip2: 'In de modus HID-Alleen is beeldmontage uitgeschakeld',
        rebuild:
          'Bij het wisselen van modus wordt de USB-verbinding opnieuw opgebouwd. IronKVM start niet opnieuw op',
        enable: 'Schakel de modus HID-Alleen in',
        disable: 'Schakel de modus HID-Alleen uit'
      },
      resetHidDone: 'USB-HID gereset',
      resetHidFailed: 'USB-HID resetten mislukt'
    },
    image: {
      driveLoaded: 'image geladen',
      driveWarning: 'bekijk de waarschuwingen',
      warning: {
        missing: 'Het imagebestand is verwijderd. De host leest de oude kopie tot u die uitwerpt.',
        writable: 'Lezen en schrijven: de host kan dit image wijzigen.',
        tooBigForCd: 'Te groot voor het cd-station ({{size}}, limiet {{max}}). Gebruik de schijf.',
        tooSmallForCd: 'Te klein voor het cd-station ({{size}}). Gebruik de schijf.',
        empty: 'Het bestand is leeg, waarschijnlijk door een mislukte upload of download.'
      },
      delete: 'Verwijderen',
      inUse: 'In gebruik. Werp het uit voordat u het verwijdert.',
      retry: 'Opnieuw',
      loadFailed: 'De lijst met images kon niet worden geladen',
      readOnlyLocked:
        'Werp de schijf uit om dit te wijzigen. Het geldt bij het plaatsen van een image.',
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
        usb1: 'Verbind de IronKVM met uw computer via USB.',
        usb2: 'Zorg ervoor dat de virtuele schijf is gekoppeld (Instellingen - Virtuele schijf).',
        usb3: 'Open de virtuele schijf op uw computer en kopieer het imagebestand naar de hoofdmap van de virtuele schijf.',
        scp1: 'Zorg ervoor dat de IronKVM en uw computer zich in hetzelfde lokale netwerk bevinden.',
        scp2: 'Open een terminal op uw computer en gebruik het SCP-commando om het imagebestand te uploaden naar de /data directory op de IronKVM.',
        scp3: 'Voorbeeld: scp uw-image-pad root@uw-nanokvm-ip:/data',
        tfCard: 'TF-kaart',
        tf1: 'Deze methode wordt ondersteund op Linux-systemen',
        tf2: 'Haal de TF-kaart uit de IronKVM (voor de VOLLEDIGE versie, demonteer eerst de behuizing).',
        tf3: 'Plaats de TF-kaart in een kaartlezer en verbind deze met uw computer.',
        tf4: 'Kopieer het imagebestand naar de /data directory op de TF-kaart.',
        tf5: 'Plaats de TF-kaart terug in de IronKVM.'
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
      close: 'Sluiten',
      empty: 'Nog geen scripts. Upload een .sh- of .py-bestand om het op het bord uit te voeren.',
      loadFailed: 'Scripts laden mislukt',
      uploaded: 'Script geüpload',
      uploadFailed: 'Script uploaden mislukt',
      started: 'Script op de achtergrond gestart',
      deleteFailed: 'Script verwijderen mislukt',
      waitLimit: 'Wachten tot het script klaar is, maximaal {{minutes}} minuten.',
      timedOut:
        'Het script liep langer dan {{minutes}} minuten en deze pagina wacht niet meer. Het draait mogelijk nog op het bord.'
    },
    terminal: {
      invalidBaud: 'Deze baudrate wordt niet ondersteund.',
      invalidPort: 'Voer een apparaatpad onder /dev in, zoals /dev/ttyS1.',
      invalidSettings:
        'Ongeldige instellingen voor de seriële poort. Dit is de shell van het bord.',
      disconnected: 'Verbinding verbroken. Druk op Enter om opnieuw te verbinden.',
      title: 'Terminal',
      nanokvm: 'IronKVM Terminal',
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
      no: 'Nee',
      yes: 'Ja',
      deleteConfirm: 'Dit opgeslagen adres verwijderen?',
      delete: 'Verwijderen',
      wake: 'Wekken',
      rename: 'Naam wijzigen',
      showMac: 'MAC-adres tonen',
      showName: 'Naam tonen',
      requestFailed: 'Het apparaat was niet bereikbaar om de opdracht te versturen',
      deleteFailed: 'Verwijderen mislukt',
      renameFailed: 'Naam wijzigen mislukt',
      title: 'Wake-on-LAN',
      sending: 'Commando wordt verzonden...',
      sent: 'Commando verzonden',
      input: 'Voer het MAC-adres in',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'Upload mislukt',
      uploadSuccess: 'Upload voltooid',
      uploading: 'Uploaden: {{file}}',
      downloadingPercent: 'Downloaden ({{percent}}): {{file}}',
      downloading: 'Downloaden: {{file}}',
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
      bootMenuPresent: '{{file}} staat al op het apparaat, met de juiste checksum',
      bootMenuDesc:
        'De netboot.xyz-ISO downloaden, met gecontroleerde checksum, voor de virtuele cd'
    },
    alerts: {
      title: 'Vraagt aandacht',
      temperature: {
        warning: 'Het bord is {{celsius}} °C. Controleer of er lucht bij kan.',
        critical: 'Het bord is {{celsius}} °C, te heet. Geef het lucht of zet het uit.'
      },
      storage: {
        warning:
          'Nog maar {{available}} van {{total}} vrij op {{path}}. Grote images passen misschien niet.',
        critical:
          'Nog maar {{available}} vrij op {{path}}. Uploads, downloads en installaties van add-ons mislukken. Verwijder images die u niet meer nodig hebt.'
      },
      vpn: '{{name}} moet bij het opstarten starten maar draait niet, dus toegang op afstand via deze VPN ligt eruit.',
      openVpn: 'VPN-instellingen openen',
      stream:
        'De videostream is mislukt. Probeer een andere videomodus in het menu Scherm of herlaad de pagina.'
    },
    power: {
      resetDesc: 'Herstart de host meteen. Niet-opgeslagen werk gaat verloren.',
      powerShortDesc: 'Zet de host aan, of vraagt het besturingssysteem af te sluiten (ACPI).',
      powerLongDesc: 'Zet de host geforceerd uit zonder af te sluiten.',
      hddLed: 'Schijf-LED',
      hddActive: 'Actief',
      hddIdle: 'Rust',
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
      nav: {
        system: 'Systeem',
        network: 'Netwerk',
        access: 'Toegang',
        integrations: 'Integraties',
        boot: 'Opstarten en media',
        browser: 'Deze browser',
        search: 'Instelling zoeken',
        noMatch: 'Geen instellingen gevonden',
        locked:
          "Er loopt een bewerking. Andere pagina's en sluiten zijn pas weer beschikbaar als die klaar is.",
        vpnProvider: 'VPN-aanbieder'
      },
      mcp: {
        keyNote:
          'MCP gebruikt een eigen API-sleutel, hieronder getoond. Sleutels van de pagina API-sleutels werken hier niet.',
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
        cancelBtn: 'Annuleren',
        showKey: 'Sleutel tonen',
        hideKey: 'Sleutel verbergen',
        regenerateKey: 'Nieuwe sleutel maken'
      },
      redfish: {
        example: 'Voorbeeld',
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
        copyBeforeSave:
          'Kopieer het wachtwoord nu. Na het opslaan kan het niet meer worden getoond.',
        noLogin:
          'IPMI staat aan, maar geen actief account heeft een IPMI-wachtwoord, dus niemand kan inloggen. Stel er hieronder een in.',
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
      ssh: {
        service: 'SSH-server',
        serviceDesc: 'sshd nu en bij elke start starten',
        failed: 'De SSH-instellingen konden niet worden geladen',
        rootDefault: 'root heeft nog het fabriekswachtwoord',
        rootEmpty: 'root heeft geen wachtwoord',
        rootWarning:
          'Iedereen die de console of SSH bereikt, kan als root inloggen. Stel een wachtwoord in onder {{account}} > {{password}}: voor de eigenaar van het apparaat stelt het ook het root-wachtwoord in.',
        connection: 'Verbinding',
        command: 'Inloggen als root',
        port: 'Poort',
        viaVpn: 'Via {{name}}',
        notRunning: 'sshd draait niet. Zet de SSH-server aan om te verbinden.',
        hostKeys: 'Vingerafdrukken van de hostsleutels',
        hostKeysDesc: 'Vergelijk ze met wat ssh bij de eerste verbinding toont.',
        noHostKeys: 'Nog geen hostsleutels. sshd maakt ze bij de eerste start.',
        keys: 'Geautoriseerde sleutels',
        keysDesc:
          'Publieke sleutels die als root kunnen inloggen. Ze staan op de datapartitie, dus updates bewaren ze.',
        noKeys: 'Nog geen geautoriseerde sleutels.',
        noComment: 'geen opmerking',
        addPlaceholder: 'Plak één publieke sleutel, zoals de inhoud van ~/.ssh/id_ed25519.pub',
        add: 'Sleutel toevoegen',
        added: 'Sleutel toegevoegd',
        removed: 'Sleutel verwijderd',
        deleteConfirm: 'Deze sleutel verwijderen?',
        deleteConfirmDesc: 'Hij kan niet meer inloggen. Open sessies blijven open.',
        invalidKey: 'Dit is geen publieke sleutel. Plak één regel uit een .pub-bestand.',
        keyOptions: 'Sleutels met opties zoals command= of from= worden hier niet geaccepteerd.',
        duplicateKey: 'Deze sleutel is al geautoriseerd.',
        lastKey: 'De laatste sleutel kan niet worden verwijderd zolang alleen-sleutels aan staat.',
        keysOnly: 'Alleen sleutels',
        keysOnlyDesc:
          'Inloggen met wachtwoord en keyboard-interactive uitzetten. Open sessies blijven open.',
        keysOnlyNeedsKey: 'Voeg eerst een geautoriseerde sleutel toe, anders kan niemand inloggen.',
        keysOnlyOn: 'Inloggen met wachtwoord uitgezet',
        keysOnlyOff: 'Inloggen met wachtwoord aangezet',
        notHonoured:
          'De sshd van deze image leest deze instelling niet, dus inloggen met wachtwoord blijft aan.',
        reloadFailed:
          'Opgeslagen, maar sshd kon niet opnieuw worden geladen. Het geldt bij de volgende start van sshd.',
        notApplied:
          'sshd accepteert nog wachtwoorden. Zet de SSH-server uit en weer aan om de instelling toe te passen.',
        changePort: 'Wijzigen',
        portConfirm: 'SSH-poort wijzigen naar {{port}}?',
        portConfirmDesc:
          'Je huidige SSH-sessies blijven open. Nieuwe verbindingen moeten poort {{port}} gebruiken. Zorg dat je firewall dat toestaat.',
        portChanged: 'SSH-poort gewijzigd naar {{port}}',
        portInvalid: 'Voer een poort van 1 tot 65535 in.',
        portReserved: 'IronKVM gebruikt deze poort zelf. Kies een andere.',
        portInUse: 'Een ander programma op de IronKVM luistert al op deze poort.',
        portNotHonoured:
          'De sshd van deze image leest deze instelling niet, dus de poort blijft ongewijzigd.'
      },
      vnc: {
        address: 'Adres',
        certHint:
          'VeNCrypt X509Plain gebruikt het zelfondertekende certificaat van het apparaat, dus de client waarschuwt bij de eerste verbinding. Accepteer het, of sla het certificaat op via het HTTPS-adres van deze pagina en geef het aan TigerVNC met -X509CA=<bestand>.',
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
        actionReset: 'Resetten',
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
      media: {
        title: 'Virtuele media',
        description:
          'Instellingen voor het venster Media in de werkbalk. Images koppelen, toevoegen en de Ventoy-set kiezen gebeurt in het venster.',
        ejectFirst:
          'De Ventoy-schijf zit in een station. Werp hem uit in het venster Media om te verwijderen.'
      },
      netboot: {
        title: 'Netwerkboot',
        isoDownload: 'Downloaden',
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
        title: 'Over IronKVM',
        information: 'Informatie',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Applicatie versie',
        applicationTip: 'Versie van de IronKVM-webapplicatie',
        image: 'Image versie',
        imageTip: 'IronKVM-kaartimage en de NanoKVM-systeemimage waarop het is gebouwd',
        kernel: 'Kernelversie',
        kernelTip: 'Versie van de Linux-kernel die nu draait',
        deviceKey: 'Apparaat sleutel',
        videoMemory: 'Videogeheugen',
        videoMemoryTip:
          'Geheugen gereserveerd voor video-opname. Het wordt niet gedeeld met de rest van het systeem.',
        videoMemoryGenerations_one: '{{count}} eerdere IronKVM-sessie houdt videogeheugen vast',
        videoMemoryGenerations_other: '{{count}} eerdere IronKVM-sessies houden videogeheugen vast',
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
        hostnameFailed: 'De hostnaam kon niet worden gewijzigd',
        editHostname: 'Hostnaam bewerken',
        docs: 'Documentatie',
        hardware: 'Hardware',
        hardwareFaq: 'Hardware-FAQ',
        disclaimer:
          'IronKVM: geharde community-firmware voor de Sipeed NanoKVM. Niet verbonden aan Sipeed.',
        basedOn: 'gebaseerd op NanoKVM {{version}}'
      },
      preferences: {
        title: 'Voorkeuren'
      },
      performance: {
        title: 'Prestaties'
      },
      appearance: {
        thisBrowser: 'Deze browser',
        thisBrowserDesc:
          'Alleen in deze browser opgeslagen. Andere browsers hebben hun eigen instellingen.',
        deviceWide: 'Apparaat',
        deviceWideDesc: 'Op het apparaat opgeslagen. Geldt voor iedereen die het opent.',
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
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Voorpaneel'
        },
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
          tip: 'Het inschakelen van deze functie kan de bruikbare levensduur van uw SD-kaart verkorten!',
          active: 'Actief - {{used}} van {{total}}',
          inactive: 'Ingesteld, maar niet in gebruik'
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
        hidOnly: 'HID-Alleen modus',
        hidOnlyDesc:
          'Stop met het emuleren van virtuele apparaten en behoud alleen de basisbesturing van HID',
        disk: 'Virtuele schijf',
        diskDesc: 'Koppel virtuele U-schijf aan de externe host',
        network: 'Virtueel Netwerk',
        networkDesc: 'Koppel virtueel netwerk kaart aan de externe host',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Host:',
          description:
            'Een privé netwerkverbinding met de externe host via de USB-kabel. De host krijgt een adres zonder gateway en zonder DNS, en kan uw LAN dus niet via IronKVM bereiken.',
          mode: 'Protocol',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (voor hosts zonder NCM)',
          rndis: 'RNDIS (niet meer aangeboden)',
          rndisNote:
            'Deze verbinding gebruikt RNDIS, dat niet meer wordt aangeboden. Kies NCM of ECM.',
          subnet: 'Subnet',
          subnetDesc:
            'Een privé IPv4-netwerk, /24 tot /30. IronKVM neemt het eerste adres, de host het tweede.',
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
          'Biedt de externe host een seriële USB-poort aan, om op deze IronKVM in te loggen als het netwerk onbereikbaar is',
        consoleTip:
          'Iedereen die de externe host bedient, krijgt een inlogprompt van deze IronKVM. Stel een sterk wachtwoord in voordat u dit inschakelt (Account - Wachtwoord wijzigen).',
        usbApply: {
          changed: 'Gewijzigd',
          discard: 'Verwerpen',
          pending: 'De wijzigingen zijn nog niet toegepast.'
        },
        endpoints: {
          title: 'USB-plaatsen',
          free: '{{free}} van {{total}} vrij',
          slots: 'Plaatsen: {{count}}',
          short: 'Niet genoeg vrije plaatsen ({{free}} vrij)',
          full: 'Niet genoeg vrije USB-plaatsen. Zet eerst iets anders uit.',
          inactive: 'Aan, maar draait niet: de USB-controller heeft geen plaatsen meer. Zet een ander apparaat uit en dit start meteen.',
          explain: 'De USB-controller heeft een vast aantal plaatsen (inkomende endpoints), en toetsenbord en muis nemen er altijd een paar. Staan er meer apparaten aan dan er passen, dan blijven toetsenbord en muis en gaat de rest uit.',
          error: 'Kan het apparaat niet bereiken. Probeer het opnieuw.',
          fitTogether: 'Deze passen samen: {{sets}}'
        },
        reboot: 'Opnieuw opstarten',
        rebootDesc: 'Weet u zeker dat u IronKVM opnieuw wilt opstarten?',
        okBtn: 'Ja',
        cancelBtn: 'Nee',
        rebootFailed: 'Herstarten mislukt'
      },
      network: {
        title: 'Netwerk',
        wifi: {
          disconnectBtn: 'Verbreken',
          disconnectWarning:
            'Als je IronKVM via dit wifi-netwerk bereikt, verliest deze pagina de verbinding.',
          disconnected: 'Wifi verbroken',
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
          waitingHttp:
            'Terugschakelen naar http. Vernieuw deze pagina als deze niet vanzelf opent.',
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
          description: 'Stel in hoe IronKVM zijn adres op het bekabelde netwerk krijgt',
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
          applyTitle: 'Het adres van IronKVM wijzigen?',
          applyWarning:
            'De verbinding met deze pagina gaat verloren. IronKVM past het nieuwe adres toe en wacht {{seconds}} seconden tot u het daar bereikt. Bereiken behoudt de wijziging. Bereikt niets het, dan zet IronKVM de vorige instellingen terug.',
          applyConfirm: 'Toepassen',
          applyCancel: 'Annuleren',
          applyFailed: 'Het adres kon niet worden toegepast',
          trialTitle: 'Wacht op bevestiging',
          trialDhcp: 'IronKVM vraagt een adres aan via DHCP.',
          trialStatic: 'IronKVM is nu bereikbaar op {{address}}.',
          trialInstruction:
            'Open IronKVM op zijn nieuwe adres en meld u aan als daarom wordt gevraagd. Daar bereiken behoudt de wijziging. Bereikt niets IronKVM binnen {{seconds}} seconden, dan zet het de vorige instellingen terug.',
          trialOpen: 'Het nieuwe adres openen',
          trialKeep: 'Deze instellingen behouden',
          trialKept: 'Het nieuwe adres is opgeslagen',
          trialKeepFailed: 'De instellingen konden niet worden behouden',
          trialGone: 'De wijziging is al teruggezet. Probeer het opnieuw.',
          unsaved: 'Niet-opgeslagen wijzigingen'
        },
        dns: {
          title: 'DNS',
          description: 'Configureer DNS-servers voor IronKVM',
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
        connect: 'Verbinden',
        connectDesc:
          'Verbind met het {{name}}-netwerk. Uit verbreekt de verbinding zonder de dienst te stoppen.',
        kvmUrl: 'KVM-adres',
        moreTip: 'Meer acties',
        restartTip: 'Herstarten',
        stopTip: 'Stoppen',
        updateTip: 'Bijwerken naar {{version}}',
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
          tip: 'Als de daemon te weinig geheugen heeft, probeer dan swap in te schakelen. Dat stelt u in bij "Instellingen > Prestaties".'
        },
        copy: 'Kopiëren',
        copied: 'Link gekopieerd',
        copyFailed: 'Link kopiëren mislukt. Selecteer hem en kopieer hem handmatig.',
        open: 'Openen',
        checkAgain: 'Opnieuw controleren',
        notSignedIn: 'Nog niet aangemeld. Rond het aanmelden via de link af en controleer opnieuw.',
        checkFailed: 'Aanmeldstatus controleren mislukt',
        loginWaiting:
          'Deze pagina controleert om de paar seconden en gaat verder zodra u bent aangemeld.',
        uninstallFailed: 'Verwijderen mislukt',
        loginFailed: 'Inloggen mislukt'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Download het',
        package: 'installatiepakket',
        unzip: 'en pak het uit',
        notLogin:
          'Het apparaat is nog niet gekoppeld. Log in en koppel dit apparaat aan uw account.',
        urlPeriod: 'Deze url is 10 minuten geldig',
        login: 'Inloggen',
        logout: 'Uitloggen',
        logoutDesc: 'Weet u zeker dat u wilt uitloggen?',
        manualIntro: 'Of installeer het handmatig via SSH:',
        copyBinaries: 'Kopieer tailscale en tailscaled naar {{dir}} op de IronKVM',
        linksFile: 'Maak in dezelfde map een bestand met de naam links met deze twee regels:',
        rebootRefresh: 'Herstart de IronKVM en vernieuw daarna deze pagina'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Dit apparaat is nog niet lid van een NetBird-netwerk. Word lid met een setup key of log in met SSO.',
        setupKey: 'Setup-sleutel',
        setupKeyPlaceholder: 'Plak een setup key uit het NetBird-dashboard',
        join: 'Lid worden',
        or: 'of',
        sso: 'Inloggen met SSO',
        urlPeriod: 'Deze url is 10 minuten geldig',
        logout: 'Afmelden',
        logoutDesc:
          'Afmelden verwijdert deze peer uit uw NetBird-account en wist de configuratie hier. Opnieuw lid worden vereist een setup key of een SSO-login, en de peer kan een nieuw IP-adres krijgen. Doorgaan?',
        joinFailed: 'Kon niet aan het netwerk deelnemen'
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
            'SHA-512 controleert alleen of het pakket overeenkomt met het manifest dat door deze server wordt verstrekt. Het bewijst niet dat het pakket een officiële IronKVM-release is. Een defecte of kwaadwillende server kan het apparaat onbruikbaar maken, gegevensverlies veroorzaken of het systeem compromitteren.',
          confirm: 'Toch gebruiken',
          useSipeed: 'De officiële Sipeed-server gebruiken',
          previewDisabled:
            'Preview-updates zijn niet beschikbaar zolang een aangepaste updateserver is ingeschakeld.'
        },
        offline: {
          chooseFile: 'Bestand kiezen',
          installing: 'Upload voltooid. Installeren...',
          noFile: 'Geen bestand gekozen',
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
          'Het apparaat installeert de update en herstart zijn server. Deze pagina herlaadt zodra de server terug is.',
        releaseNotes: 'Release-opmerkingen'
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
        mcpNote:
          'Deze sleutels werken niet voor MCP, dat een eigen sleutel heeft op de MCP-pagina.',
        metricsUrl: 'Metrics-URL',
        monitoring: 'Monitoring',
        monitoringDesc:
          'Prometheus leest de metrics met een API-sleutel van deze pagina, verzonden als Bearer-token. Elke rol mag ze lezen.',
        scrapeConfig: 'Prometheus-scrapeconfiguratie',
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
        kvmDescription: 'Bedien de externe host via IronKVM.',
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
          restoring: 'PicoClaw herstellen',
          ready: 'Runtime gereed',
          stopped: 'Runtime gestopt',
          blockedByMCP: 'Externe MCP-bediening is actief',
          readyBlockedByMCP:
            'De runtime draait, maar externe MCP bedient nu de invoer van het apparaat.',
          readyWithoutControl:
            'De runtime draait. Geef PicoClaw apparaatbediening voordat je opnieuw verbindt.',
          unavailable: 'Runtime niet beschikbaar',
          configError: 'Configuratiefout'
        },
        transport: {
          connecting: 'Verbinden',
          connected: 'Verbonden',
          disconnected: 'Verbinding verbroken',
          reconnect: 'Opnieuw verbinden',
          reconnectDescription: 'Opnieuw verbinden met de lopende PicoClaw-sessie.',
          reconnectBlocked: 'PicoClaw heeft apparaatbediening nodig om opnieuw te verbinden.'
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
        picoclawDescription:
          'PicoClaw kan toetsenbord- en muisinvoer sturen. Handmatige invoer kan pauzeren.',
        mcp: 'Apparaatbediening: externe MCP',
        mcpDescription:
          'Externe MCP kan naar het apparaat schrijven. PicoClaw neemt de invoer niet over.',
        off: 'Apparaatbediening: uit',
        offDescription:
          'De AI stuurt geen toetsenbord- of muisinvoer. Handmatige bediening blijft beschikbaar.',
        transitioning: 'Apparaatbediening: wisselen',
        transitioningDescription: 'De apparaatbediening wordt gesynchroniseerd. Even geduld.',
        grant: 'Bediening geven',
        release: 'Vrijgeven',
        releasing: 'Vrijgeven...',
        switching: 'Wisselen...',
        releasingLabel: 'Apparaatbediening: vrijgeven',
        releasingDescription:
          'De apparaatbediening wordt teruggegeven. PicoClaw is gestopt met lopende invoer.',
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
        switchFromMCP: 'Overschakelen naar PicoClaw en starten',
        takeoverAndStart: 'Overnemen en starten'
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
    upstream: {
      check: 'Controleren op updates',
      updateTo: 'Bijwerken naar {{version}}',
      confirm: '{{name}} bijwerken naar {{version}}?',
      confirmDesc:
        'De nieuwe release wordt van GitHub gedownload en gecontroleerd met de checksums die erbij gepubliceerd zijn. Als er iets misgaat, blijft de huidige versie staan.',
      ok: 'Bijwerken',
      upToDate: 'Up-to-date',
      builtIn: 'ingebouwd',
      checkFailed: 'Kon niet op updates controleren: {{error}}',
      unverifiable: 'Versie {{version}} wordt niet aangeboden: {{reason}}',
      inUse: 'Bijwerken kan nu niet: {{reason}}',
      running: 'Bijwerken naar {{version}}...',
      done: '{{name}} bijgewerkt naar {{version}}',
      failed: 'De laatste update is mislukt: {{error}}'
    },
    menu: {
      mediaAdd: 'Image toevoegen',
      mediaMoreOptions: 'Meer opties',
      mediaSettings: 'Media-instellingen',
      collapse: 'Menu samenvouwen',
      expand: 'Menu uitvouwen',
      more: 'Meer',
      media: 'Media',
      tools: 'Hulpmiddelen',
      text: 'Tekst',
      advanced: 'Geavanceerd',
      mediaMounted: 'Gekoppeld',
      mediaLibrary: 'Bibliotheek',
      textToHost: 'Naar de host',
      textFromHost: 'Van de host'
    },
    ion: {
      checking: 'Videogeheugen controleren voordat de stream start...',
      warn: 'Het videogeheugen raakt op. Eén herstart van de server zou het uitputten. Start opnieuw op wanneer het u uitkomt.',
      criticalTitle: 'Niet genoeg videogeheugen om de stream te starten',
      criticalBody:
        'Video starten zou het gereserveerde geheugen uitputten en de server stoppen. Alle andere functies blijven werken, inclusief stroombeheer en opnieuw opstarten. Alleen opnieuw opstarten van de IronKVM maakt dit geheugen vrij.',
      criticalContinue: 'Video toch starten',
      criticalReboot: 'IronKVM opnieuw opstarten',
      criticalRebooting: 'Opnieuw opstarten...'
    }
  }
};

export default nl;
