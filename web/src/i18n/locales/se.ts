const se = {
  translation: {
    head: {
      desktop: 'Fjärrskrivbord',
      login: 'Logga in',
      changePassword: 'Byt lösenord',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        'Webbläsaren vägrade spara sessionen. En cookie från en tidigare HTTPS-session kan inte ersättas över vanlig http. Rensa cookies för den här adressen, eller öppna ett privat fönster, och logga in igen.',
      login: 'Logga in',
      placeholderUsername: 'Användarnamn',
      placeholderPassword: 'Lösenord',
      placeholderCurrentPassword: 'Nuvarande lösenord',
      placeholderPassword2: 'Vänligen ange lösenordet igen',
      noEmptyUsername: 'Användarnamn krävs',
      noEmptyPassword: 'Lösenord krävs',
      passwordLength: 'Lösenordet måste vara mellan 8 och 72 tecken',
      noAccount: 'Kunde inte hämta användarinformation, uppdatera sidan eller återställ lösenordet',
      invalidUser: 'Ogiltigt användarnamn eller lösenord',
      locked: 'För många inloggningar, försök igen senare',
      globalLocked: 'System under skydd, försök igen senare',
      error: 'Oväntat fel',
      invalidCurrentPassword: 'Nuvarande lösenord är felaktigt',
      changePassword: 'Byt lösenord',
      changePasswordDesc: 'För din enhets säkerhet, byt lösenord!',
      differentPassword: 'Lösenorden matchar inte',
      illegalUsername: 'Användarnamnet innehåller ogiltiga tecken',
      illegalPassword: 'Lösenordet innehåller ogiltiga tecken',
      forgetPassword: 'Glömt lösenord',
      ok: 'Ok',
      cancel: 'Avbryt',
      loginButtonText: 'Logga in',
      tips: {
        reset1: 'För att återställa lösenordet, håll in BOOT-knappen på NanoKVM i 10 sekunder.',
        reset2: 'För detaljerade steg, se detta dokument:',
        reset3: 'Standardkonto för webben:',
        reset4: 'Standardkonto för SSH:',
        change1: 'Observera att denna åtgärd ändrar följande lösenord:',
        change2: 'Webbinloggningslösenord',
        change3: 'Systemets root-lösenord (SSH-lösenord)',
        change4: 'För att återställa lösenordet, håll in BOOT-knappen på NanoKVM.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Konfigurera Wi-Fi för NanoKVM',
      success: 'Kontrollera nätverksstatusen för NanoKVM och besök den nya IP-adressen.',
      failed: 'Åtgärden misslyckades, försök igen.',
      invalidMode:
        'Det aktuella läget stöder inte nätverksinstallation. Gå till din enhet och aktivera Wi-Fi konfigurationsläge.',
      confirmBtn: 'Ok',
      finishBtn: 'Färdig',
      ap: {
        authTitle: 'Autentisering krävs',
        authDescription: 'Ange lösenordet AP för att fortsätta',
        authFailed: 'Ogiltigt AP lösenord',
        passPlaceholder: 'AP lösenord',
        verifyBtn: 'Verifiera'
      }
    },
    screen: {
      scale: 'Skala',
      title: 'Skärm',
      video: 'Videoläge',
      videoDirectTips: 'Aktivera HTTPS i "Inställningar > Enhet" för att använda detta läge',
      resolution: 'Upplösning',
      ocr: {
        title: 'Läs text (OCR)',
        tips: 'Texten tolkas i den här webbläsaren. Du kan rätta den innan du kopierar den.',
        hint: 'Dra över texten som ska läsas. Tryck på Esc för att avbryta.',
        noPicture: 'Vänta på videon och dra sedan över texten som ska läsas.',
        cancel: 'Avbryt',
        language: 'Språk',
        languages: {
          eng: 'Engelska'
        },
        preview: 'Markerat område',
        capturing: 'Fångar skärmen...',
        loading: 'Läser in textigenkänning...',
        recognizing: 'Läser texten...',
        noText: 'Ingen text hittades i det markerade området.',
        copy: 'Kopiera',
        copied: 'Kopierat till urklipp',
        copyFailed: 'Det gick inte att kopiera till urklipp',
        selectAgain: 'Markera igen',
        unsupported:
          'Den här webbläsaren kan inte köra textigenkänning. Den kräver WebAssembly SIMD, som dagens webbläsare har.',
        captureFailed: 'Det gick inte att fånga skärmen.',
        outside: 'Det markerade området ligger utanför bilden.',
        recognizeFailed: 'Textigenkänningen misslyckades.'
      },
      controlRegion: {
        title: 'Muskalibrering',
        description:
          'Använd den här inställningen när den styrda enheten använder en upplösning som inte är 16:9 och markören är feljusterad i sid- eller höjdled.',
        off: 'Av',
        auto: 'Automatisk',
        autoWarning: 'Kalibreringen kan misslyckas om användarprogrammet har en helsvart bakgrund.',
        manual: 'Manuell',
        selectedResolution: 'Valt områdesupplösning',
        unused: 'Används inte',
        originalResolution: 'Ursprunglig upplösning',
        selectResolution: 'Välj ursprunglig upplösning',
        addResolution: 'Lägg till anpassad upplösning',
        add: 'Lägg till',
        duplicateResolution: 'Den här upplösningen finns redan.',
        width: 'Bredd',
        height: 'Höjd',
        apply: 'Beräkna och tillämpa',
        invalidResolution: 'Ange en giltig ursprunglig upplösning när videon är klar.',
        select: 'Välj område',
        clear: 'Återställ automatiskt',
        saveFailed: 'Det gick inte att spara inmatningsområdet.',
        tooSmall: 'Det valda området är för litet.',
        previewUnavailable: 'Förhandsvisning är inte tillgänglig',
        clearConfirm: 'Återställa automatisk identifiering av svarta kanter?',
        dragHint: 'Dra för att välja området för fjärrskrivbordet',
        finish: 'Klar',
        confirm: 'Bekräfta',
        cancel: 'Avbryt'
      },
      auto: 'Automatisk',
      autoTips:
        'Skärmtear eller musförskjutning kan förekomma vid vissa upplösningar. Överväg att justera fjärrvärdens upplösning eller inaktivera automatiskt läge.',
      fps: 'FPS',
      customizeFps: 'Anpassa',
      quality: 'Kvalitet',
      qualityLossless: 'Förlustfri',
      qualityHigh: 'Hög',
      qualityMedium: 'Medel',
      qualityLow: 'Låg',
      frameDetect: 'Ramdetection',
      frameDetectTip:
        'Beräkna skillnaden mellan ramar. Sluta skicka videoström när inga förändringar upptäcks på fjärrvärdens skärm.',
      resetHdmi: 'Återställ HDMI',
      mixedH264: {
        title: 'H.264-strömningskonflikt',
        description:
          'H.264 Direct och H.264 WebRTC används samtidigt. Detta kan orsaka skärmrivningar eller skadad video. Använd endast ett H.264-läge.'
      },
      webrtcConnectionFailed: {
        title: 'WebRTC-anslutningen misslyckades',
        description: 'Kontrollera nätverksanslutningen eller byt videoläge.'
      },
      captureStatus: {
        hdmiError: 'HDMI-skärmfel',
        unsupportedResolution: 'Den aktuella upplösningen stöds inte',
        retrieving: 'Hämtar skärmen...',
        changingResolution: 'Byter upplösning...',
        updateFailed: 'Skärmen kan inte uppdateras just nu',
        videoError: 'Videovisningsfel',
        noHdmi: 'Ingen HDMI-signal upptäcktes',
        unavailable: 'Skärmen kan inte visas just nu'
      }
    },
    keyboard: {
      title: 'Tangentbord',
      paste: 'Klistra in',
      tips: 'Skriver texten på värden som tangenttryckningar. Välj den tangentbordslayout som värden använder.',
      placeholder: 'Ange text',
      submit: 'Skicka',
      virtual: 'Tangentbord',
      readClipboard: 'Läs från Urklipp',
      clipboardPermissionDenied:
        'Behörighet till Urklipp nekad. Vänligen tillåt åtkomst till Urklipp i din webbläsare.',
      clipboardReadError: 'Misslyckades med att läsa Urklipp',
      mediaKeys: {
        title: 'Medieknappar',
        mute: 'Ljud av',
        volumeDown: 'Sänk volymen',
        volumeUp: 'Höj volymen',
        previous: 'Föregående spår',
        playPause: 'Spela upp eller pausa',
        next: 'Nästa spår',
        stop: 'Stopp'
      },
      pasting: {
        layout: 'Tangentbordslayout på värden',
        layouts: {
          us: 'Engelska (USA)',
          uk: 'Engelska (Storbritannien)',
          de: 'Tyska',
          fr: 'Franska',
          es: 'Spanska',
          it: 'Italienska',
          ptBr: 'Portugisiska (Brasilien)',
          se: 'Svenska / finska',
          ru: 'Ryska',
          ja: 'Japanska',
          ko: 'Koreanska'
        },
        speed: 'Skrivhastighet',
        speeds: {
          fast: 'Snabb',
          normal: 'Normal',
          slow: 'Långsam'
        },
        estimate: 'Skrivtid: ungefär {{duration}}',
        untypeable: 'Tecken som den här layouten inte kan skriva: {{count}}',
        untypeableAt: 'rad {{line}}, kolumn {{column}}',
        skipUntypeable: 'Skriv resten',
        shortcut: '{{shortcut}} skriver urklippet på värden direkt.',
        clipboardUnavailable:
          'Webbläsaren låter bara en sida läsa urklippet över HTTPS. Klistra in texten i rutan med Ctrl+V.',
        clipboardEmpty: 'Urklippet innehåller ingen text.',
        tooLong: 'Texten är för lång. Gränsen är {{max}} tecken.',
        inProgress: 'En inklistring skrivs redan.',
        typing: 'Skriver på värden',
        done: 'Text skriven',
        canceled: 'Inklistring avbruten',
        failed: 'Inklistring misslyckades',
        cancel: 'Avbryt',
        controlBusy: 'En annan styrning använder tangentbordet.',
        hidError: 'Tangenttryckningarna kunde inte skickas till värden.'
      },
      shortcut: {
        title: 'Genvägar',
        custom: 'Anpassad',
        capture: 'Klicka här för att fånga genväg',
        clear: 'Rensa',
        save: 'Spara',
        captureTips:
          'Att fånga systemtangenter (som Windows-tangenten) kräver helskärmsbehörighet.',
        enterFullScreen: 'Växla helskärmsläge.'
      },
      leaderKey: {
        title: 'Leader-tangent',
        desc: 'Gå förbi webbläsarbegränsningar och skicka systemgenvägar direkt till fjärrvärden.',
        howToUse: 'Hur man använder',
        simultaneous: {
          title: 'Samtidigt läge',
          desc1: 'Håll ned Leader-tangenten och tryck sedan på genvägen.',
          desc2: 'Intuitivt, men kan komma i konflikt med systemgenvägar.'
        },
        sequential: {
          title: 'Sekventiellt läge',
          desc1:
            'Tryck på Leader-tangenten → tryck på genvägen i följd → tryck på Leader-tangenten igen.',
          desc2: 'Kräver fler steg, men undviker helt systemkonflikter.'
        },
        enable: 'Aktivera Leader-tangent',
        tip: 'När den tilldelas som Leader-tangent fungerar denna tangent endast som genvägsutlösare och förlorar sitt standardbeteende.',
        placeholder: 'Tryck på Leader-tangenten',
        shiftRight: 'Höger Shift',
        ctrlRight: 'Höger Ctrl',
        metaRight: 'Höger Win',
        submit: 'Skicka',
        recorder: {
          rec: 'REC',
          activate: 'Aktivera tangenter',
          input: 'Vänligen tryck på genvägen...'
        }
      }
    },
    mouse: {
      title: 'Mus',
      cursor: 'Markörstil',
      default: 'Standardmarkör',
      pointer: 'Pekmarkör',
      cell: 'Cellmarkör',
      text: 'Textmarkör',
      grab: 'Greppmarkör',
      hide: 'Dölj markör',
      mode: 'Musläge',
      absolute: 'Absolut läge',
      relative: 'Relativt läge',
      absoluteShort: 'Absolut',
      relativeShort: 'Relativ',
      absoluteStalled: 'Målet ignorerar den absoluta musen',
      absoluteStalledDesc:
        'Målet har slutat hämta absoluta musrapporter, så pekarens rörelser går förlorade. Tangentbordet påverkas inte. Att återställa USB brukar lösa det; relativt läge använder en annan slutpunkt.',
      useRelative: 'Byt till relativt läge',
      direction: 'Rullhjulsriktning',
      scrollUp: 'Scrolla uppåt',
      scrollDown: 'Scrolla ner',
      speed: 'Rullhjulshastighet',
      fast: 'Snabb',
      slow: 'Långsam',
      requestPointer: 'Använder relativt läge. Klicka på skrivbordet för att få muspekaren.',
      resetHid: 'Återställ HID',
      hidOnly: {
        title: 'Endast HID-läge',
        desc: 'Om din mus och ditt tangentbord slutar svara och återställning av HID inte hjälper, kan det bero på kompatibilitetsproblem mellan NanoKVM och enheten. Prova att aktivera Endast-HID-läge för bättre kompatibilitet.',
        tip1: 'Aktivering av Endast-HID-läge avmonterar den virtuella U-disken och nätverket',
        tip2: 'I Endast-HID-läge är avbildningsmontering inaktiverat',
        rebuild: 'Byte av läge bygger upp USB-anslutningen på nytt. NanoKVM startas inte om',
        enable: 'Aktivera Endast-HID-läge',
        disable: 'Inaktivera Endast-HID-läge'
      }
    },
    image: {
      title: 'Avbildningar',
      loading: 'Laddar...',
      empty: 'Inget hittades',
      mountMode: 'Monteringsläge',
      mountFailed: 'Montering misslyckades',
      mountDesc:
        'I vissa system måste den virtuella disken avmonteras på fjärrvärden innan avbildningen monteras.',
      unmountFailed: 'Avmontering misslyckades',
      unmountDesc:
        'I vissa system måste du manuellt mata ut från fjärrvärden innan du avmonterar avbildningen.',
      refresh: 'Uppdatera avbildningslistan',
      disk: 'Disk',
      cdrom: 'CD',
      driveEmpty: 'Tom',
      eject: 'Mata ut',
      readOnly: 'Skrivskyddad',
      readOnlyTip: 'Gäller nästa avbildning som sätts in i disken.',
      noDrives: 'Inga virtuella enheter. Aktivera den virtuella disken i Inställningar.',
      insertFailed: 'Isättning misslyckades',
      ejectFailed: 'Utmatning misslyckades',
      insertInto: 'Sätts in i {{drive}}. Klicka för att ändra.',
      loadedIn: 'Isatt i {{drive}}',
      attention: 'Observera',
      deleteConfirm: 'Är du säker på att du vill ta bort denna avbildning?',
      okBtn: 'Ja',
      cancelBtn: 'Nej',
      tips: {
        title: 'Hur man laddar upp',
        usb1: 'Anslut NanoKVM till din dator via USB.',
        usb2: 'Säkerställ att den virtuella disken är monterad (Inställningar - Virtuell Disk).',
        usb3: 'Öppna den virtuella disken på din dator och kopiera avbildningsfilen till rotkatalogen.',
        scp1: 'Säkerställ att NanoKVM och din dator är på samma lokala nätverk.',
        scp2: 'Öppna en terminal på din dator och använd SCP-kommandot för att ladda upp avbildningen till /data på NanoKVM.',
        scp3: 'Exempel: scp din-avbildningssökväg root@din-nanokvm-ip:/data',
        tfCard: 'TF-kort',
        tf1: 'Denna metod stöds på Linux-system',
        tf2: 'Ta ut TF-kortet från NanoKVM (för FULL-versionen, öppna chassit först).',
        tf3: 'Sätt in TF-kortet i en kortläsare och anslut till din dator.',
        tf4: 'Kopiera avbildningsfilen till /data på TF-kortet.',
        tf5: 'Sätt in TF-kortet i NanoKVM.'
      }
    },
    script: {
      title: 'Skript',
      upload: 'Ladda upp',
      run: 'Kör',
      runBackground: 'Kör i bakgrunden',
      runFailed: 'Körning misslyckades',
      attention: 'Observera',
      delDesc: 'Är du säker på att du vill ta bort denna fil?',
      confirm: 'Ja',
      cancel: 'Nej',
      delete: 'Ta bort',
      close: 'Stäng'
    },
    terminal: {
      title: 'Terminal',
      nanokvm: 'NanoKVM Terminal',
      serial: 'Serieport-terminal',
      serialPort: 'Serieport',
      serialPortPlaceholder: 'Ange serieport',
      baudrate: 'Baudhastighet',
      parity: 'Paritet',
      parityNone: 'Ingen',
      parityEven: 'Jämn',
      parityOdd: 'Udda',
      flowControl: 'Flödeskontroll',
      flowControlNone: 'Ingen',
      flowControlSoft: 'Programvara',
      flowControlHard: 'Hårdvara',
      dataBits: 'Databitar',
      stopBits: 'Stoppbitar',
      confirm: 'Ok'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'Skickar kommando...',
      sent: 'Kommando skickat',
      input: 'Ange MAC-adress',
      ok: 'Ok'
    },
    download: {
      title: 'Avbildningshämtare',
      input: 'Ange en fjärravbildnings-URL',
      ok: 'Ok',
      disabled: '/data partitionen är skrivskyddad, kan inte hämta avbildning',
      uploadbox: 'Släpp filen här eller klicka för att välja',
      inputfile: 'Vänligen ange bildfilen',
      NoISO: 'Ingen ISO',
      sha256: 'SHA-256 (valfrie)',
      sha256Placeholder: 'Skriv inn en SHA-256-kontrollsum på 64 tegn',
      invalidSHA256: 'SHA-256 må være en heksadesimal streng på 64 tegn',
      failed: 'Nedlasting mislyktes',
      success: 'Nedlasting fullført',
      checksumFailed: 'Nedlasting mislyktes: SHA-256-verifisering mislyktes',
      cancel: 'Avbryt',
      cancelFailed: 'Kunne ikke avbryte nedlastingen',
      bootMenu: 'Startmeny (netboot.xyz)',
      bootMenuDesc:
        'Ladda ner netboot.xyz-ISO:n, med kontrollerad kontrollsumma, till den virtuella cd:n'
    },
    power: {
      title: 'Ström',
      showConfirm: 'Bekräftelse',
      showConfirmTip: 'Strömätgärder kräver extra bekräftelse',
      reset: 'Starta om',
      power: 'Ström',
      powerShort: 'Ström (kort tryck)',
      powerLong: 'Ström (långt tryck)',
      resetConfirm: 'Utföra omstart?',
      powerConfirm: 'Utföra strömåtgärd?',
      okBtn: 'Ja',
      cancelBtn: 'Nej',
      hostOs: 'Värdens OS',
      hostOsTip: 'Skickas som USB-tangenter. Värden bestämmer vad de gör.',
      sleep: 'Viloläge',
      wake: 'Väck',
      wakeKey: 'Väck med Shift',
      powerDown: 'Stäng av',
      sleepConfirm: 'Försätta värden i viloläge?',
      powerDownConfirm: 'Skicka avstängningstangenten till värden?',
      wakeTip:
        'En värd i viloläge ignorerar ofta Väck från enheten som försatte den i viloläge. Väck med Shift trycker på en tangent på tangentbordet, vilket fler värdar godtar.',
      led: 'Ström-LED',
      ledOn: 'På',
      ledOff: 'Av',
      ledUnknown: 'Okänd',
      ledConnected: 'Ström-LED ansluten',
      ledConnectedTip:
        'Aktivera endast om värdens stiftlist för ström-LED är kopplad till kortet. Utan den är strömläget okänt.',
      ledConnectedFailed: 'Det gick inte att spara inställningen för ström-LED'
    },
    settings: {
      title: 'Inställningar',
      mcp: {
        title: 'MCP-tjänst',
        service: 'MCP-fjärrstyrning',
        serviceDesc:
          'Tillåt betrodda MCP-klienter att styra tangentbord och mus och ta skärmbilder',
        securityWarning:
          'Alla som har denna API-nyckel kan styra fjärrvärden och se dess skärm. Använd HTTPS och aktivera tjänsten endast i betrodda nätverk.',
        endpoint: 'Slutpunkt',
        apiKey: 'API-nyckel',
        regenerateConfirmTitle: 'Generera om MCP API-nyckeln?',
        regenerateConfirmDesc: 'Den aktuella nyckeln slutar omedelbart att fungera.',
        enableConfirmTitle: 'Aktivera extern MCP-styrning?',
        enableConfirmDesc:
          'Om MCP aktiveras stoppas PicoClaw och alla aktiva PicoClaw-sessioner stängs.',
        failed: 'MCP-åtgärden misslyckades',
        copyFailed: 'Kopiering misslyckades. Kopiera manuellt.',
        okBtn: 'Bekräfta',
        cancelBtn: 'Avbryt'
      },
      redfish: {
        title: 'Redfish',
        service: 'Redfish-tjänst',
        serviceDesc:
          'DMTF Redfish API, för strömstyrning, virtuella medier och status från verktyg som redfishtool och Ansible. Om den stängs av avslutas alla Redfish-sessioner.',
        endpoint: 'Tjänsterot',
        httpsOn: 'Kortet använder HTTPS, vilket de flesta Redfish-verktyg kräver.',
        httpsOff:
          'Kortet använder vanlig HTTP. De flesta Redfish-verktyg kräver HTTPS: aktivera det i "Inställningar > Nätverk".',
        credentials:
          'Redfish använder KVM-kontona, med Basic-autentisering eller en Redfish-session, och API-nycklar som skickas som X-Auth-Token. API-nycklar hanteras på sidan API-nycklar.',
        powerActions: 'Strömåtgärder',
        powerActionsDesc:
          'De återställningstyper som erbjuds just nu. On, ForceOff och GracefulShutdown kräver strömläget, så de erbjuds bara när "Ström-LED ansluten" är aktiverat i strömmenyn.',
        sessions: 'Sessioner',
        noSessions: 'Inga öppna Redfish-sessioner',
        created: 'Skapad',
        lastUsed: 'Senast använd',
        refresh: 'Uppdatera',
        end: 'Avsluta',
        endConfirmTitle: 'Avsluta den här Redfish-sessionen?',
        endConfirmDesc: 'Dess token slutar fungera direkt. Klienten måste logga in igen.',
        failed: 'Redfish-åtgärden misslyckades',
        copyFailed: 'Kopieringen misslyckades. Kopiera manuellt.',
        okBtn: 'Bekräfta',
        cancelBtn: 'Avbryt'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Värd-watchdog',
        serviceDesc:
          'Om värden borde vara igång och dess bild inte ändras, eller det saknas HDMI-signal, under hela tidsgränsen, trycker kortet på reset eller stänger av och slår på värden.',
        stillWarning:
          'En värd vars skärm går i viloläge, eller vars bild står still medan den arbetar, ser ut att ha hängt sig. Stäng av skärmvila på värden, eller ange en ping-adress.',
        ledHint:
          '"Ström-LED ansluten" är avstängt i strömmenyn. Watchdogen ser inte när värden är avstängd, så den behandlar värden som alltid på.',
        timeout: 'Tidsgräns',
        timeoutDesc: 'Hur länge värden får vara utan livstecken innan watchdogen ingriper.',
        action: 'Åtgärd',
        actionDesc:
          'Av och på håller strömknappen intryckt i 5 sekunder och trycker sedan på den igen.',
        actionReset: 'Reset',
        actionPower: 'Av och på',
        cooldown: 'Paus',
        cooldownDesc: 'Den kortaste tiden mellan två åtgärder.',
        maxPerHour: 'Åtgärder per timme',
        maxPerHourDesc: 'Det största antalet åtgärder under en timme.',
        pingHost: 'Ping-adress',
        pingHostDesc:
          'Värdens IP-adress. Ett svar räknas som ett livstecken. Lämna tomt för att inte pinga.',
        pingHostInvalid: 'Ange en IPv4- eller IPv6-adress.',
        minutes: 'min',
        save: 'Spara',
        saved: 'Sparat',
        state: 'Detektor',
        status: {
          off: 'Av',
          watching: 'Övervakar',
          hostOff: 'Värd avstängd',
          captureOff: 'HDMI-inspelning av',
          cooldown: 'Paus',
          capped: 'Timgräns nådd',
          acting: 'Ingriper'
        },
        signal: 'HDMI-signal',
        yes: 'Ja',
        no: 'Nej',
        led: 'Ström-LED',
        on: 'Tänd',
        off: 'Släckt',
        ledNotConnected: 'Inte ansluten',
        ping: 'Ping',
        pingNotSet: 'Inte angiven',
        pingReply: 'Svarar',
        pingNoReply: 'Inget svar',
        lastChange: 'Senaste bildändring',
        never: 'Aldrig',
        actsIn: 'Ingriper om',
        actionsLastHour: 'Åtgärder senaste timmen',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Logg',
        noLog: 'Watchdogen har inte ingripit än.',
        refresh: 'Uppdatera',
        reasonFrozen: 'Bilden ändrades inte',
        reasonNoSignal: 'Ingen HDMI-signal',
        stuckFor: 'inga livstecken på {{duration}}',
        pressFailed: 'Knapptrycket misslyckades: {{error}}',
        noScreenshot: 'Ingen skärmbild',
        failed: 'Watchdog-åtgärden misslyckades'
      },
      netboot: {
        title: 'Nätverksstart',
        description:
          'Starta värden från nätverket: iPXE och en meny med avbildningarna på KVM:en över USB-nätverkslänken, eller netboot.xyz via proxy-DHCP på LAN:et.',
        addon: 'dnsmasq och startfiler',
        addonDesc:
          'Installerade på /data: dnsmasq från Alpine, iPXE och netboot.xyz från deras utgåvor, var och en kontrollerad mot sin kontrollsumma.',
        install: 'Installera',
        installing: 'Installerar. Det kan ta några minuter.',
        uninstall: 'Avinstallera',
        uninstallConfirm: 'Stäng av nätverksstart och ta bort dnsmasq och startfilerna?',
        needsData: 'Nätverksstart kräver en IronKVM-avbildning med /data-partitionen monterad.',
        usb: 'På USB-nätverkslänken',
        usbDesc:
          'Medan USB-nätverkslänken är på betjänar dnsmasq den i stället för udhcpd. Värden får sin enda adress utan router och utan DNS-server, iPXE för sin arkitektur och en meny med ISO-avbildningarna på KVM:en.',
        linkOff: 'USB-nätverkslänken är av. Slå på den under Enhet, USB-nätverk.',
        menuUrl: 'Meny',
        leases: 'Värdens lease',
        noLeases: 'Inga ännu',
        netbootxyzNote:
          'netboot.xyz i menyn laddas från internet, som USB-länken inte når. Värden behöver internet på en annan nätverksport.',
        lan: 'Proxy-DHCP på LAN:et',
        lanDesc:
          'Svarar PXE-klienter på LAN:et med netboot.xyz, som sedan laddar sin meny från internet. Den delar aldrig ut adresser och erbjuder inte avbildningarna på KVM:en.',
        lanWarning:
          'Alla PXE-klienter på detta LAN erbjuds netboot.xyz, inte bara värden. Slå på detta bara i ett nätverk du själv styr.',
        lanConfirm: 'Slå på proxy-DHCP på LAN:et?',
        lanInterface: 'LAN',
        running: 'Körs',
        stopped: 'Körs inte',
        images: 'Avbildningar i menyn',
        noImages: 'Inga ISO-avbildningar i avbildningskatalogen.',
        boots: 'Senaste starter',
        noBoots: 'Värden har inte hämtat något ännu.',
        log: 'dnsmasq-logg',
        refresh: 'Uppdatera',
        okBtn: 'Bekräfta',
        cancelBtn: 'Avbryt',
        failed: 'Nätverksstartsåtgärden misslyckades'
      },
      about: {
        title: 'Om NanoKVM',
        information: 'Information',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Applikationsversion',
        applicationTip: 'NanoKVM webbapplikationsversion',
        image: 'Systemversion',
        imageTip: 'NanoKVM systemavbildningsversion',
        kernel: 'Kärnversion',
        kernelTip: 'Version av Linux-kärnan som körs nu',
        deviceKey: 'Enhetsnyckel',
        videoMemory: 'Videominne',
        videoMemoryTip:
          'Minne reserverat för videoinspelning. Det delas inte med resten av systemet.',
        videoMemoryGenerations_one: '{{count}} tidigare NanoKVM-session håller videominne',
        videoMemoryGenerations_other: '{{count}} tidigare NanoKVM-sessioner håller videominne',
        videoMemoryReboot: 'Starta om för att frigöra det.',
        community: 'Community',
        hostname: 'Värdnamn',
        hostnameUpdated: 'Värdnamn uppdaterat. Starta om för att tillämpa.',
        ipType: {
          Wired: 'Trådbundet',
          Wireless: 'Trådlöst',
          Other: 'Annat'
        }
      },
      appearance: {
        title: 'Utseende',
        display: 'Visning',
        language: 'Språk',
        languageDesc: 'Välj språk för gränssnittet',
        webTitle: 'Webbtitel',
        webTitleDesc: 'Anpassa webbsidans titel',
        menuBar: {
          title: 'Menyrad',
          mode: 'Visningsläge',
          modeDesc: 'Visa menyraden på skärmen',
          modeOff: 'Av',
          modeAuto: 'Dölj automatiskt',
          modeAlways: 'Alltid synlig',
          keyboardLedStatus: 'Indikatorer för tangentbordslås',
          keyboardLedStatusDesc:
            'Visa Num Lock-, Caps Lock- och Scroll Lock-status för fjärrdatorn',
          icons: 'Undermenyikoner',
          iconsDesc: 'Visa undermenyikoner i menyraden'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Status för lås på fjärrtangentbord',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'På',
        off: 'Av',
        unknown: 'Okänd'
      },
      device: {
        title: 'Enhet',
        oled: {
          title: 'OLED',
          description: 'Stäng av OLED-skärmen efter',
          brightness: 'OLED-ljusstyrka',
          brightnessDescription: 'En lägre nivå gör att skärmen håller längre',
          brightnessLevels: {
            '64': 'Lägst',
            '96': 'Låg',
            '128': 'Medel',
            '160': 'Hög',
            '207': 'Standard',
            '255': 'Max'
          },
          0: 'Aldrig',
          15: '15 sek',
          30: '30 sek',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 timme'
        },
        ssh: {
          description: 'Aktivera SSH-fjärråtkomst',
          tip: 'Ställ in ett starkt lösenord innan du aktiverar (Konto - Byt lösenord)'
        },
        advanced: 'Avancerade inställningar',
        cpuFreq: {
          title: 'CPU-frekvens',
          description: 'Ange CPU-klockan som används vid nästa start',
          tip: 'CPU:n startar på 850 MHz och är specificerad för 1000 MHz. Ett nytt värde tillämpas vid nästa start, inte medan systemet körs. 1000 MHz ligger inom specifikationen; temperaturen håller sig väl inom gränserna med båda inställningarna.',
          running: 'Körs: {{mhz}} MHz',
          rebootToApply: 'starta om för att tillämpa',
          rebootConfirm: 'Starta om nu för att tillämpa {{mhz}} MHz?'
        },
        swap: {
          title: 'Swap',
          disable: 'Inaktivera',
          description: 'Ange swap-filens storlek',
          tip: 'Aktivering av denna funktion kan förkorta livslängden på ditt SD-kort!'
        },
        zram: {
          title: 'Komprimerad swap (zram)',
          description: 'Swap i komprimerat RAM i stället för på SD-kortet',
          tip: 'zram håller swap borta från SD-kortet, så det orsakar inget slitage. Det finns ingen disk-swap bakom: om zram blir fullt stoppar kärnan en process i stället för att växla långsamt. Minnesgränsen begränsar hur mycket RAM zram kan ta.',
          unavailable: 'Kärnmodulerna är inte installerade på den här enheten',
          inactive: 'Aktiverad, men enheten startade inte',
          active: 'Aktiv - {{used}} av {{total}}, {{ratio}}x',
          off: 'Av',
          detail: {
            algorithm: 'Algoritm: {{algorithm}}',
            memory: 'Använt minne: {{used}} av {{limit}}',
            memoryNoLimit: 'Använt minne: {{used}}, ingen gräns satt',
            counters: 'Sidor växlade in {{in}}, ut {{out}} (alla swap-enheter, sedan start)'
          }
        },
        mouseJiggler: {
          title: 'Musvickare',
          description: 'Förhindra att fjärrvärden går i viloläge',
          disable: 'Inaktivera',
          absolute: 'Absolut läge',
          relative: 'Relativt läge'
        },
        mdns: {
          description: 'Aktivera mDNS-upptäckningstjänst',
          tip: 'Stäng av om det inte behövs'
        },
        hdmi: {
          description: 'Aktivera HDMI/monitorutgång',
          idleTimeoutTitle: 'Tidsgräns för inaktiv inspelning',
          idleTimeoutDescription:
            'Stoppa HDMI-inspelning efter att det inte har funnits aktiva tittare i',
          minutes: 'min'
        },
        autostart: {
          title: 'Autostart skriptinställningar',
          description: 'Hantera skript som körs automatiskt vid systemstart',
          new: 'Nytt',
          deleteConfirm: 'Är du säker på att du vill ta bort denna fil?',
          yes: 'Ja',
          no: 'Nej',
          scriptName: 'Autostart skriptnamn',
          scriptContent: 'Autostart skriptinnehåll',
          settings: 'Inställningar'
        },
        hidOnly: 'Endast-HID-läge',
        hidOnlyDesc: 'Sluta emulera virtuella enheter, behåll bara grundläggande HID kontroll',
        disk: 'Virtuell disk',
        diskDesc: 'Montera virtuell U-disk på fjärrvärden',
        network: 'Virtuellt nätverk',
        networkDesc: 'Montera virtuell nätverkskort på fjärrvärden',
        usbNetwork: {
          description:
            'En privat nätverkslänk till fjärrvärden via USB-kabeln. Värden får en adress utan gateway och utan DNS, så den kan inte nå ditt lokala nätverk genom NanoKVM.',
          off: 'Av',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (för värdar utan NCM)',
          rndis: 'RNDIS (erbjuds inte längre)',
          rndisNote: 'Den här länken använder RNDIS, som inte längre erbjuds. Välj NCM eller ECM.',
          subnet: 'Delnät',
          subnetDesc:
            'Ett privat IPv4-nätverk, /24 till /30. NanoKVM tar den första adressen, värden den andra.',
          addresses: 'NanoKVM: {{board}}, värd: {{host}}',
          invalidSubnet: 'Ange ett delnät, till exempel 172.31.255.0/30.',
          apply: 'Verkställ',
          confirm: 'Återansluta USB-enheten?',
          reenumerate:
            'När du verkställer byggs USB-anslutningen upp på nytt. Värden förlorar tangentbord, mus och virtuell disk i några sekunder.'
        },
        audio: 'Virtuell högtalare',
        audioDesc:
          'Visa ett USB-ljudkort för fjärrvärden, så att du kan höra den. Värden måste välja det som sin utenhet. Att ändra detta bygger upp USB-anslutningen på nytt.',
        audioNote: 'Ljud finns i båda H.264-lägena (WebRTC och Direct), inte i MJPEG',
        console: 'Seriell konsol',
        consoleDesc:
          'Visa en seriell USB-port för fjärrvärden, för att logga in på denna NanoKVM när nätverket inte går att nå',
        consoleTip:
          'Alla som styr fjärrvärden får en inloggningsprompt till denna NanoKVM. Ställ in ett starkt lösenord innan du aktiverar (Konto - Byt lösenord).',
        endpoints: {
          title: 'USB-slutpunkter',
          used: '{{used}} av {{total}} används',
          cost: 'använder {{cost}}',
          needs: 'behöver {{cost}}',
          full: 'Inte tillräckligt med USB-slutpunkter. Stäng av något annat först.',
          inactive:
            'På, men körs inte: USB-styrenheten fick slut på slutpunkter. Stäng av en annan enhet så startar den här direkt.',
          explain:
            'USB-styrenheten har ett fast antal inkommande slutpunkter, och det är dem som räknas här. Om fler enheter är aktiverade än det finns plats för behålls tangentbord och mus, och resten stängs av.',
          error: 'Kunde inte nå enheten. Försök igen.',
          fitTogether: 'Dessa ryms tillsammans: {{sets}}'
        },
        reboot: 'Starta om',
        rebootDesc: 'Är du säker på att du vill starta om NanoKVM?',
        okBtn: 'Ja',
        cancelBtn: 'Nej'
      },
      network: {
        title: 'Nätverk',
        wifi: {
          title: 'Wi-Fi',
          description: 'Konfigurera Wi-Fi',
          apMode: 'AP-läge är aktiverat, anslut till Wi-Fi genom att skanna QR-koden',
          connect: 'Anslut Wi-Fi',
          connectDesc1: 'Ange nätverkets SSID och lösenord',
          connectDesc2: 'Ange lösenordet för att ansluta till detta nätverk',
          disconnect: 'Är du säker på att du vill koppla från nätverket?',
          failed: 'Anslutningen misslyckades, försök igen.',
          ssid: 'Namn',
          password: 'Lösenord',
          joinBtn: 'Anslut',
          confirmBtn: 'OK',
          cancelBtn: 'Avbryt'
        },
        tls: {
          description: 'Aktivera HTTPS-protokoll',
          tip: 'Observera: Användning av HTTPS kan öka fördröjningen, särskilt med MJPEG-läge.',
          restarting: 'Startar om enhetens server, det tar ungefär två minuter...',
          waiting: 'Väntar på att enheten svarar igen...',
          waitingHttp: 'Byter tillbaka till http. Ladda om sidan om den inte öppnas av sig själv.'
        },
        ethernet: {
          title: 'IP-adress',
          description: 'Ställ in hur NanoKVM får sin adress i det trådbundna nätverket',
          dhcp: 'DHCP',
          manual: 'Manuell',
          networkDetails: 'Nätverksinformation',
          interface: 'Gränssnitt',
          ipAddress: 'IP-adress',
          subnetMask: 'Nätmask',
          router: 'Router',
          save: 'Tillämpa',
          invalidAddress: 'Ange en giltig IP-adress',
          invalidMask: 'Ange en giltig nätmask, till exempel 255.255.255.0 eller 24',
          invalidRouter: 'Ange en giltig routeradress',
          addressRequired: 'En IP-adress krävs',
          maskRequired: 'En nätmask krävs',
          applyTitle: 'Vill du ändra adressen för NanoKVM?',
          applyWarning:
            'Anslutningen till den här sidan går förlorad. NanoKVM tillämpar den nya adressen och väntar {{seconds}} sekunder på att du når den där. Att nå den behåller ändringen. Om ingenting når den återställer NanoKVM de tidigare inställningarna.',
          applyConfirm: 'Tillämpa',
          applyCancel: 'Avbryt',
          applyFailed: 'Adressen kunde inte tillämpas',
          trialTitle: 'Väntar på bekräftelse',
          trialDhcp: 'NanoKVM begär en adress via DHCP.',
          trialStatic: 'NanoKVM finns nu på {{address}}.',
          trialInstruction:
            'Öppna NanoKVM på dess nya adress och logga in om den ber om det. Att nå den där behåller ändringen. Om ingenting når NanoKVM inom {{seconds}} sekunder återställs de tidigare inställningarna.',
          trialOpen: 'Öppna den nya adressen',
          trialKeep: 'Behåll dessa inställningar',
          trialKept: 'Den nya adressen är sparad',
          trialKeepFailed: 'Inställningarna kunde inte behållas',
          trialGone: 'Ändringen har redan återställts. Försök igen.',
          unsaved: 'Osparade ändringar'
        },
        dns: {
          title: 'DNS',
          description: 'Konfigurera DNS-servrar för NanoKVM',
          mode: 'Läge',
          dhcp: 'DHCP',
          manual: 'Manuell',
          add: 'Lägg till DNS',
          save: 'Spara',
          invalid: 'Ange en giltig IP-adress',
          noDhcp: 'Ingen DHCP-DNS är tillgänglig just nu',
          saved: 'DNS-inställningar sparade',
          saveFailed: 'Det gick inte att spara DNS-inställningar',
          unsaved: 'Osparade ändringar',
          maxServers: 'Maximalt {{count}} DNS-servrar tillåtna',
          dnsServers: 'DNS-servrar',
          dhcpServersDescription: 'DNS-servrar hämtas automatiskt från DHCP',
          manualServersDescription: 'DNS-servrar kan redigeras manuellt',
          networkDetails: 'Nätverksdetaljer',
          interface: 'Gränssnitt',
          ipAddress: 'IP-adress',
          subnetMask: 'Subnätmask',
          router: 'Router',
          none: 'Ingen'
        }
      },
      vpn: {
        loading: 'Laddar...',
        okBtn: 'Ja',
        cancelBtn: 'Nej',
        restart: 'Starta om {{name}}?',
        stop: 'Stoppa {{name}}?',
        stopDesc:
          'Tjänsten stoppas nu. Starta vid uppstart är en separat inställning och förblir som den är.',
        update: 'Uppdatera {{name}} till {{version}}?',
        updateDesc: 'Tjänsten startas om om den körs. Inloggningen behålls.',
        notInstall: '{{name}} är inte installerad.',
        install: 'Installera',
        installing: 'Installerar',
        installFailed: 'Installationen misslyckades',
        retry: 'Försök igen',
        notRunning: '{{name}} körs inte. Starta den för att fortsätta.',
        run: 'Starta',
        boot: 'Starta vid uppstart',
        bootDesc: 'Starta {{name}} när KVM:en startar.',
        enable: 'Aktivera {{name}}',
        control: 'Kontrollserver',
        connected: 'Ansluten',
        disconnected: 'Inte ansluten',
        deviceName: 'Enhetsnamn',
        deviceIP: 'Enhetens IP',
        account: 'Konto',
        version: 'Version',
        uptime: 'Drifttid',
        peers: 'Noder',
        noPeers: 'Inga noder ännu.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Minne',
        daemonRss: 'Tjänst',
        group: 'Tilläggsgrupp',
        high: 'stryps över {{size}}',
        max: 'stoppas av kärnan över {{size}}',
        noGroup: 'Ingen minnesgrupp för tillägg på det här kortet.',
        uninstall: 'Avinstallera {{name}}',
        uninstallDesc:
          'Är du säker på att du vill avinstallera {{name}}? Inloggningen blir kvar på kortet.',
        blocked:
          '{{other}} körs eller startar vid uppstart. Bara ett VPN kan köras åt gången: stoppa {{other}} och stäng av dess start vid uppstart först.',
        swap: {
          title: 'Swap-minne',
          tip: 'Om tjänsten får ont om minne kan du prova att aktivera swap-minne. Detta sätter swap-filens storlek till 256MB som standard, vilket kan justeras i "Inställningar > Enhet".'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Uppdatera sidan och försök igen. Eller installera manuellt',
        download: 'Ladda ner',
        package: 'installationspaketet',
        unzip: 'och packa upp det',
        upTailscale: 'Ladda upp tailscale till NanoKVM-katalogen /usr/bin/',
        upTailscaled: 'Ladda upp tailscaled till NanoKVM-katalogen /usr/sbin/',
        refresh: 'Uppdatera sidan',
        notLogin: 'Enheten är ännu inte bunden. Logga in och bind enheten till ditt konto.',
        urlPeriod: 'Denna URL är giltig i 10 minuter',
        login: 'Logga in',
        loginSuccess: 'Inloggning lyckades',
        logout: 'Logga ut',
        logoutDesc: 'Är du säker på att du vill logga ut?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Den här enheten har inte anslutit till ett NetBird-nätverk ännu. Anslut med en installationsnyckel, eller logga in med SSO.',
        setupKey: 'Installationsnyckel',
        setupKeyPlaceholder: 'Klistra in en installationsnyckel från NetBird-instrumentpanelen',
        join: 'Anslut',
        or: 'eller',
        sso: 'Logga in med SSO',
        urlPeriod: 'Denna URL är giltig i 10 minuter',
        loginSuccess: 'Inloggning lyckades',
        logout: 'Avregistrera',
        logoutDesc:
          'Avregistrering tar bort den här noden från ditt NetBird-konto och raderar dess konfiguration här. För att ansluta igen krävs en installationsnyckel eller en SSO-inloggning, och noden kan få en ny IP. Fortsätta?'
      },
      update: {
        title: 'Sök efter uppdateringar',
        queryFailed: 'Kunde inte hämta version',
        updateFailed: 'Uppdatering misslyckades. Försök igen.',
        isLatest: 'Du har redan den senaste versionen.',
        available: 'En uppdatering finns tillgänglig. Vill du uppdatera nu?',
        updating: 'Uppdatering påbörjad. Vänta...',
        confirm: 'Bekräfta',
        cancel: 'Avbryt',
        preview: 'Förhandsvisning av uppdateringar',
        previewDesc: 'Få tidig tillgång till nya funktioner och förbättringar',
        previewTip:
          'Observera att förhandsversioner kan innehålla buggar eller ofullständig funktionalitet!',
        customServer: {
          title: 'Anpassad uppdateringsserver',
          desc: 'Sök efter och hämta onlineuppdateringar från en angiven server',
          invalidUrl:
            'Ange en giltig HTTP- eller HTTPS-serverkatalog utan frågesträng, fragment eller latest.json.',
          loadFailed: 'Det gick inte att läsa in uppdateringsserverns konfiguration.',
          saveFailed: 'Det gick inte att spara uppdateringsserverns konfiguration.',
          saved: 'Uppdateringsserverns konfiguration har sparats.',
          save: 'Spara',
          confirmTitle: 'Vill du använda en anpassad uppdateringsserver?',
          confirmDesc:
            'SHA-512 kontrollerar endast att paketet överensstämmer med manifestet från den här servern. Det bevisar inte att paketet är en officiell NanoKVM-utgåva. En felaktig eller skadlig server kan göra enheten obrukbar, orsaka dataförlust eller äventyra systemets säkerhet.',
          confirm: 'Använd ändå',
          useSipeed: 'Använd Sipeeds officiella server',
          previewDisabled:
            'Förhandsuppdateringar är inte tillgängliga när en anpassad uppdateringsserver är aktiverad.'
        },
        offline: {
          title: 'Offlineuppdateringar',
          desc: 'Uppdatera genom lokalt installationspaket',
          upload: 'Ladda upp',
          checksumPlaceholder: 'SHA-256-kontrollsumma (valfri)',
          invalidChecksum: 'SHA-256-kontrollsumman måste innehålla 64 hexadecimala tecken.',
          checksumMismatch: 'SHA-256-verifieringen misslyckades. Paketet kan vara skadat.',
          invalidName: 'Ogiltigt filnamnsformat. Ladda ner från GitHub-versioner.',
          updateFailed: 'Uppdatering misslyckades. Försök igen.'
        }
      },
      account: {
        title: 'Konto',
        webAccount: 'Webbkonto-namn',
        role: 'Roll',
        roles: { admin: 'Administratör', user: 'Användare' },
        password: 'Lösenord',
        updateBtn: 'Byt',
        logoutBtn: 'Logga ut',
        logoutDesc: 'Är du säker på att du vill logga ut?',
        okBtn: 'Ja',
        cancelBtn: 'Nej',
        users: {
          title: 'Användare',
          create: 'Skapa användare',
          enabled: 'Aktiverad',
          disabled: 'Inaktiverad',
          deviceOwner: 'Enhetens ägare',
          resetPassword: 'Återställ lösenord',
          delete: 'Ta bort',
          deleteConfirm: 'Ta bort den här användaren och återkalla alla dess sessioner?',
          created: 'Användaren skapades',
          deleted: 'Användaren togs bort',
          passwordUpdated: 'Lösenordet uppdaterades',
          loadFailed: 'Det gick inte att läsa in användare',
          saveFailed: 'Det gick inte att spara användaren',
          deleteFailed: 'Det gick inte att ta bort användaren'
        }
      },
      apiKeys: {
        title: 'API-nycklar',
        description:
          'En nyckel agerar som sin ägare, med den användarens roll. Skicka den som Authorization: Bearer <key> för mätvärden och API:t, eller som X-Auth-Token för Redfish.',
        name: 'Namn',
        namePlaceholder: 'Vad nyckeln är till för, till exempel prometheus',
        nameRequired: 'Ge nyckeln ett namn',
        nameTooLong: 'Namnet får vara högst 64 tecken',
        unnamed: '(namnlös)',
        create: 'Skapa nyckel',
        created: 'Skapad',
        owner: 'Ägare',
        empty: 'Inga API-nycklar',
        newKeyTitle: 'Din nya API-nyckel',
        newKeyWarning:
          'Kopiera nyckeln nu. Den sparas inte och kan inte visas igen. Om du tappar bort den, återkalla den och skapa en ny.',
        copy: 'Kopiera',
        copied: 'Kopierad',
        copyFailed: 'Kopieringen misslyckades. Kopiera manuellt.',
        done: 'Klar',
        revoke: 'Återkalla',
        revokeConfirmTitle: 'Återkalla den här API-nyckeln?',
        revokeConfirmDesc: 'Allt som använder "{{name}}" slutar fungera direkt.',
        revoked: 'API-nyckeln återkallades',
        loadFailed: 'Det gick inte att läsa in API-nycklar',
        createFailed: 'Det gick inte att skapa API-nyckeln',
        revokeFailed: 'Det gick inte att återkalla API-nyckeln',
        cancelBtn: 'Avbryt'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistent',
      empty: 'Öppna panelen och starta en uppgift för att börja.',
      inputPlaceholder: 'Beskriv vad du vill att PicoClaw ska göra',
      newConversation: 'Ny konversation',
      processing: 'Bearbetar...',
      agent: {
        defaultTitle: 'Allmän assistent',
        defaultDescription: 'Allmän hjälp för chatt, sökning och arbetsyta.',
        kvmTitle: 'Fjärrstyrning',
        kvmDescription: 'Manövrera fjärrvärden genom NanoKVM.',
        switched: 'Agentroll bytte',
        switchFailed: 'Det gick inte att byta agentroll'
      },
      send: 'Skicka',
      cancel: 'Avbryt',
      status: {
        connecting: 'Ansluter till gateway...',
        connected: 'PicoClaw-session ansluten',
        disconnected: 'PicoClaw-session frånkopplad',
        stopped: 'Stoppbegäran har skickats',
        runtimeStarted: 'PicoClaw runtime startad',
        runtimeStartFailed: 'Det gick inte att starta PicoClaw runtime',
        runtimeStopped: 'PicoClaw runtime stoppad',
        runtimeStopFailed: 'Det gick inte att stoppa PicoClaw runtime',
        controlSwitchedToMCP: 'Styrningen har växlats till den externa MCP-tjänsten'
      },
      connection: {
        runtime: {
          checking: 'Kontrollerar',
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime klar',
          stopped: 'Runtime stoppad',
          blockedByMCP: 'Extern MCP-styrning är aktiv',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime inte tillgänglig',
          configError: 'Konfigurationsfel'
        },
        transport: {
          connecting: 'Ansluter',
          connected: 'Ansluten',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'Inaktiv',
          busy: 'Upptagen'
        }
      },
      message: {
        toolAction: 'Åtgärd',
        observation: 'Observation',
        screenshot: 'Skärmdump'
      },
      overlay: {
        locked: 'PicoClaw styr enheten. Manuell inmatning är pausad.'
      },
      control: {
        picoclaw: 'Enhetsstyrning: PicoClaw',
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Enhetsstyrning: extern MCP',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Enhetsstyrning: av',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Ge styrning',
        release: 'Släpp',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'PicoClaw-styrning beviljad',
        released: 'PicoClaw-styrning släppt',
        grantFailed: 'Det gick inte att ge PicoClaw styrning',
        releaseFailed: 'Det gick inte att släppa PicoClaw styrning',
        grantConfirmTitle: 'Växla enhetsstyrning till PicoClaw?',
        grantConfirmDesc: 'Externa MCP-enhetsskrivningar kommer att avbrytas.'
      },
      install: {
        install: 'Installera PicoClaw',
        installing: 'Installerar PicoClaw',
        success: 'PicoClaw har installerats framgångsrikt',
        failed: 'Det gick inte att installera PicoClaw',
        uninstalling: 'Avinstallerar runtime...',
        uninstalled: 'Runtime avinstallerades framgångsrikt.',
        uninstallFailed: 'Avinstallationen misslyckades.',
        requiredTitle: 'PicoClaw är inte installerad',
        requiredDescription: 'Installera PicoClaw innan du startar PicoClaw runtime.',
        progressDescription: 'PicoClaw laddas ner och installeras.',
        stages: {
          preparing: 'Förbereder',
          downloading: 'Laddar ner',
          extracting: 'Packar upp',
          verifying: 'Verifierar',
          installing: 'Installerar',
          installed: 'Installerad',
          install_timeout: 'Timeout',
          install_failed: 'Misslyckades'
        }
      },
      model: {
        requiredTitle: 'Modellkonfiguration krävs',
        requiredDescription: 'Konfigurera PicoClaw-modellen innan du använder PicoClaw-chatten.',
        docsTitle: 'Konfigurationsguide',
        docsDesc: 'Modeller och protokoll som stöds',
        menuLabel: 'Konfigurera modell',
        modelIdentifier: 'Modellidentifierare',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API-nyckel',
        apiKeyPlaceholder: 'Ange modellens API-nyckel',
        save: 'Spara',
        saving: 'Sparar',
        saved: 'Modellkonfiguration sparad',
        saveFailed: 'Det gick inte att spara modellkonfigurationen',
        invalid: 'Modellidentifierare, API Base URL och API-nyckel krävs'
      },
      uninstall: {
        menuLabel: 'Avinstallera',
        confirmTitle: 'Avinstallera PicoClaw',
        confirmContent:
          'Är du säker på att du vill avinstallera PicoClaw? Detta kommer att radera den körbara filen och alla konfigurationsfiler.',
        confirmOk: 'Avinstallera',
        confirmCancel: 'Avbryt'
      },
      history: {
        title: 'Historik',
        loading: 'Laddar sessioner...',
        emptyTitle: 'Ingen historik än',
        emptyDescription: 'Tidigare PicoClaw-sessioner kommer att visas här.',
        loadFailed: 'Det gick inte att ladda sessionshistoriken',
        deleteFailed: 'Det gick inte att ta bort sessionen',
        deleteConfirmTitle: 'Ta bort session',
        deleteConfirmContent: 'Är du säker på att du vill ta bort "{{title}}"?',
        deleteConfirmOk: 'Ta bort',
        deleteConfirmCancel: 'Avbryt',
        messageCount_one: '{{count}} meddelande',
        messageCount_other: '{{count}} meddelanden',
        messageCount: '{{count}} meddelanden'
      },
      config: {
        startRuntime: 'Starta PicoClaw',
        stopRuntime: 'Stoppa PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Växla styrningen till PicoClaw?',
        enableConfirmDesc: 'När PicoClaw startas inaktiveras den externa MCP-tjänsten.',
        enableConfirmOk: 'Starta PicoClaw',
        enableConfirmCancel: 'Avbryt',
        title: 'Starta PicoClaw',
        description: 'Starta runtime för att börja använda PicoClaw-assistenten.',
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'Vi stötte på ett problem',
      refresh: 'Uppdatera',
      panel: 'Den här delen av sidan slutade fungera',
      retry: 'Försök igen'
    },
    fullscreen: {
      toggle: 'Växla fullskärm'
    },
    input: {
      disconnected: 'Tangentbord och mus är inte anslutna',
      disconnectedTls:
        'Webbläsaren avvisade den säkra anslutningen som bär tangentbord och mus, och det gör den utan att fråga. Certifikatet som den här enheten skapade är inte betrott ännu. Öppna den här adressen i en ny flik, godkänn certifikatet och ladda sedan om. Att installera certifikatet är den pålitliga lösningen.',
      disconnectedNever:
        'Anslutningen som bär tangentbord och mus kunde inte öppnas. Resten av sidan fungerar eftersom den inte använder den. Kontrollera att inget mellan dig och enheten blockerar den.',
      disconnectedDropped:
        'Anslutningen som bär tangentbord och mus bröts och har inte kommit tillbaka. Den återansluter av sig själv efter en omstart; om detta kvarstår, ladda om sidan.',
      hidDisabled: 'HID är avstängt på den här enheten (/boot/disable_hid).',
      keyFailed: 'Tangenten kunde inte skickas.'
    },
    speaker: { title: 'Högtalare', unmute: 'Slå på ljud', mute: 'Stäng av ljud' },
    menu: {
      collapse: 'Fäll ihop menyn',
      expand: 'Expandera menyn'
    },
    ion: {
      checking: 'Kontrollerar videominnet innan strömmen startar...',
      warn: 'Lite videominne kvar. En omstart av servern skulle ta slut på det. Starta om när det passar.',
      criticalTitle: 'Inte tillräckligt med videominne för att starta strömmen',
      criticalBody:
        'Att starta video skulle ta slut på det reserverade minnet och stoppa servern. Alla andra funktioner fungerar fortfarande, även strömstyrning och omstart. Endast en omstart av NanoKVM frigör detta minne.',
      criticalContinue: 'Starta video ändå',
      criticalReboot: 'Starta om NanoKVM',
      criticalRebooting: 'Startar om...'
    }
  }
};

export default se;
