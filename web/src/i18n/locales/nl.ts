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
      login: 'Inloggen',
      placeholderUsername: 'Voer gebruikersnaam in',
      placeholderPassword: 'Voer wachtwoord in',
      placeholderPassword2: 'Voer wachtwoord nogmaals in',
      noEmptyUsername: 'Gebruikersnaam mag niet leeg zijn',
      noEmptyPassword: 'Wachtwoord mag niet leeg zijn',
      noAccount:
        'Ophalen van gebruikersinformatie mislukt, vernieuw de webpagina of reset het wachtwoord',
      invalidUser: 'Ongeldige gebruikersnaam of wachtwoord',
      locked: 'Te veel aanmeldingen, probeer het later opnieuw',
      globalLocked: 'Systeem wordt beveiligd. Probeer het later opnieuw',
      error: 'Onverwachte fout',
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
        tip3: 'NanoKVM wordt automatisch opnieuw opgestart na het wisselen van modus',
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
      attention: 'Let op',
      deleteConfirm: 'Weet u zeker dat u deze afbeelding wilt verwijderen?',
      okBtn: 'Ja',
      cancelBtn: 'Nee',
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
      showConfirmTip: 'Stroombedieningen vereisen een extra bevestiging',
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
        'Een slapende host negeert Wekken vaak van het apparaat dat hem in slaap bracht. Wekken met Shift drukt een toets op het toetsenbord in, die meer hosts accepteren.'
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
      watchdog: {
        title: 'Watchdog',
        service: 'Host-watchdog',
        serviceDesc:
          'Als de host aan zou moeten staan en zijn beeld niet verandert, of er geen HDMI-signaal is, gedurende de time-out, drukt het bord op reset of zet het de host uit en weer aan.',
        stillWarning:
          'Een host waarvan het scherm in slaapstand gaat, of waarvan het beeld stilstaat terwijl hij werkt, lijkt vastgelopen. Zet de schermslaapstand op de host uit, of stel een ping-adres in.',
        ledHint:
          '"Power LED connected" staat uit in het stroommenu. De watchdog ziet niet wanneer de host uit staat, dus behandelt hij de host als altijd aan.',
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
        failed: 'Watchdog-bewerking mislukt'
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
        deviceKey: 'Apparaat sleutel',
        community: 'Community',
        hostname: 'Hostnaam',
        hostnameUpdated: 'Hostnaam bijgewerkt. Start opnieuw op om toe te passen.',
        ipType: {
          Wired: 'Bedraad',
          Wireless: 'Draadloos',
          Other: 'Anders'
        }
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
        swap: {
          title: 'Wisselen',
          disable: 'Uitschakelen',
          description: 'Stel de grootte van het wisselbestand in',
          tip: 'Het inschakelen van deze functie kan de bruikbare levensduur van uw SD-kaart verkorten!'
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
        endpoints: {
          fitTogether: 'Deze passen samen: {{sets}}'
        },
        reboot: 'Opnieuw opstarten',
        rebootDesc: 'Weet u zeker dat u NanoKVM opnieuw wilt opstarten?',
        okBtn: 'Ja',
        cancelBtn: 'Nee'
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
          tip: 'Let op: HTTPS gebruiken kan de latentie verhogen, vooral in MJPEG-videomodus.'
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
      tailscale: {
        title: 'Tailscale',
        memory: {
          title: 'Geheugen optimalisatie',
          tip: 'Wanneer geheugen gebruik de limiet overschreid, garbage collection wordt agressiever uitgevoerd om geheugen vrij te maken. geadviseerd om 50MB te kiezen als Tailscale wordt gebruikt. Tailscale moet worden herstart om de wijziging door te voeren.'
        },
        swap: {
          title: 'Geheugen wisselen',
          tip: 'Als de problemen aanhouden nadat u geheugenoptimalisatie hebt ingeschakeld, probeer dan het wisselgeheugen in te schakelen. Hierdoor wordt de grootte van het wisselbestand standaard ingesteld op 256MB, wat kan worden aangepast in "Instellingen > Apparaat".'
        },
        restart: 'Weet u zeker dat u Tailscale opnieuw wilt opstarten?',
        stop: 'Weet u zeker dat u Tailscale wilt stoppen?',
        stopDesc: 'Meld Tailscale af en schakel het automatisch opstarten bij het opstarten uit.',
        loading: 'Laden...',
        notInstall: 'Tailscale niet gevonden! Installeer a.u.b.',
        install: 'Installeren',
        installing: 'Installeren bezig',
        failed: 'Installatie mislukt',
        retry: 'Vernieuw en probeer opnieuw. Of probeer handmatig te installeren',
        download: 'Download het',
        package: 'installatiepakket',
        unzip: 'en pak het uit',
        upTailscale: 'Upload tailscale naar NanoKVM directory /usr/bin/',
        upTailscaled: 'Upload tailscaled naar NanoKVM directory /usr/sbin/',
        refresh: 'Vernieuw huidige pagina',
        notRunning: 'Tailscale is niet actief. Start het programma om door te gaan.',
        run: 'Begin',
        notLogin:
          'Het apparaat is nog niet gekoppeld. Log in en koppel dit apparaat aan uw account.',
        urlPeriod: 'Deze url is 10 minuten geldig',
        login: 'Inloggen',
        loginSuccess: 'Inloggen gelukt',
        enable: 'Tailscale inschakelen',
        deviceName: 'Apparaatnaam',
        deviceIP: 'Apparaat IP',
        account: 'Account',
        logout: 'Uitloggen',
        logoutDesc: 'Weet u zeker dat u wilt uitloggen?',
        uninstall: 'Verwijderen Tailscale',
        uninstallDesc: 'Weet u zeker dat u Tailscale wilt verwijderen?',
        okBtn: 'Ja',
        cancelBtn: 'Nee'
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
        }
      },
      account: {
        title: 'Account',
        webAccount: 'Web Account Naam',
        password: 'Wachtwoord',
        updateBtn: 'Update',
        logoutBtn: 'Afmelden',
        logoutDesc: 'Weet u zeker dat u wilt uitloggen?',
        okBtn: 'Ja',
        cancelBtn: 'Nee'
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
      refresh: 'Vernieuwen'
    },
    fullscreen: {
      toggle: 'Volledig scherm schakelen'
    },
    input: {
      hidDisabled: 'HID is uitgeschakeld op dit apparaat (/boot/disable_hid).',
      keyFailed: 'De toets kon niet worden verzonden.'
    },
    menu: {
      collapse: 'Menu samenvouwen',
      expand: 'Menu uitvouwen'
    }
  }
};

export default nl;
