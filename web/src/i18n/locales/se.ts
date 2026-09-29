const se = {
  translation: {
    feedback: {
      enabled: '{{name}} aktiverat',
      disabled: '{{name}} inaktiverat',
      failed: 'Begäran misslyckades. Försök igen.',
      network: 'Kunde inte nå enheten. Kontrollera anslutningen och försök igen.',
      saved: 'Sparat',
      timeout: 'Enheten tog för lång tid att svara. Försök igen.'
    },
    common: {
      copy: 'Kopiera',
      copied: 'Kopierat',
      copyFailed: 'Kunde inte kopiera. Markera texten och kopiera den manuellt.',
      notUpdating: 'Uppdateras inte: den senaste uppdateringen misslyckades.',
      off: 'Av',
      running: 'Körs',
      save: 'Spara',
      cancel: 'Avbryt',
      delete: 'Ta bort',
      remove: 'Ta bort'
    },
    head: {
      desktop: 'Fjärrskrivbord',
      login: 'Logga in',
      changePassword: 'Byt lösenord',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Lösenordet har ändrats. Logga in med det nya lösenordet.',
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
        reset1: 'För att återställa lösenordet, håll in BOOT-knappen på IronKVM i 10 sekunder.',
        reset3: 'Standardkonto för webben:',
        reset4: 'Standardkonto för SSH:',
        change1: 'Observera att denna åtgärd ändrar följande lösenord:',
        change2: 'Webbinloggningslösenord',
        change3: 'Systemets root-lösenord (SSH-lösenord)',
        change4: 'För att återställa lösenordet, håll in BOOT-knappen på IronKVM.',
        resetDocs: 'Detaljerade steg finns i hårdvarudokumentationen:',
        hardwareDocs: 'Sipeed NanoKVM-wiki'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Konfigurera Wi-Fi för IronKVM',
      success: 'Kontrollera nätverksstatusen för IronKVM och besök den nya IP-adressen.',
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
      },
      ssidRequired: 'Ange nätverksnamnet, högst 32 tecken',
      passwordLength: 'Lösenordet är 8 till 63 tecken. Lämna det tomt för ett öppet nätverk.',
      passwordOptional: 'Lösenord (tomt för ett öppet nätverk)',
      lost: 'Kortet slutade svara. Det kan ha anslutit till nätverket och stängt sin konfigurationshotspot. Om hotspoten kommer tillbaka misslyckades anslutningen: anslut till den igen och försök på nytt.',
      done: 'Konfigurationen är klar. Anslut den här enheten till ditt vanliga nätverk igen och öppna kortet på dess nya adress.'
    },
    screen: {
      viewOnly: 'Endast visning',
      viewOnlyTip:
        'Den här fliken slutar skicka tangentbord och mus till värden. Skript, musjigglern och andra tittare påverkas inte.',
      viewOnlyOff: 'Stäng av endast visning',
      viewOnlyBlocked: 'Endast visning är på, så inget skickades till värden',
      pauseHidden: 'Pausa när fliken är dold',
      pauseHiddenTip:
        'Stoppar video och ljud några sekunder efter att fliken döljs och startar dem igen när du kommer tillbaka.',
      screenshot: 'Skärmbild',
      screenshotTip: 'Sparar värdens skärm som PNG i full inspelningsstorlek.',
      screenshotFailed: 'Skärmbilden misslyckades',
      stream: {
        ok: 'bild OK',
        noSignal: 'ingen signal',
        failed: 'strömmen misslyckades'
      },
      codecNoWebrtcHevc: 'Den här webbläsaren kan inte ta emot H.265 via WebRTC',
      codecNoHevc: 'Den här webbläsaren kan inte avkoda H.265',
      codecNote:
        'Kortet har en kodare, så detta ändrar strömmen för alla tittare. Anslut igen för att tillämpa det på en pågående WebRTC-session.',
      codec: 'Kodek',
      updateFailed: 'Inställningen tillämpades inte',
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
      qualityLossless: 'Bäst',
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
      },
      directConnectionFailed: 'Anslutningen till videoströmmen misslyckades'
    },
    keyboard: {
      close: 'Stäng',
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
        sendFailed: 'Inte skickat: indataanslutningen är nere',
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
        saveFailed: 'Det gick inte att spara ledartangenten',
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
      jiggler: 'Musrörare',
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
      touch: 'Pekskärmsläge',
      touchShort: 'Pekskärm',
      absoluteStalled: 'Målet ignorerar den absoluta musen',
      absoluteStalledDesc:
        'Målet har slutat hämta absoluta musrapporter, så pekarens rörelser går förlorade. Tangentbordet påverkas inte. Att återställa USB brukar lösa det; relativt läge använder en annan slutpunkt.',
      useRelative: 'Byt till relativt läge',
      direction: 'Rullhjulsriktning',
      scrollUp: 'Som på den här datorn',
      scrollDown: 'Omvänd (naturlig rullning)',
      speed: 'Rullhjulshastighet',
      fast: 'Snabb',
      slow: 'Långsam',
      requestPointer: 'Använder relativt läge. Klicka på skrivbordet för att få muspekaren.',
      resetHid: 'Återställ HID',
      hidOnly: {
        switchFailed: 'Det gick inte att byta läge. Kontrollera anslutningen och försök igen.',
        title: 'Endast HID-läge',
        desc: 'Om din mus och ditt tangentbord slutar svara och återställning av HID inte hjälper, kan det bero på kompatibilitetsproblem mellan IronKVM och enheten. Prova att aktivera Endast-HID-läge för bättre kompatibilitet.',
        tip1: 'Aktivering av Endast-HID-läge avmonterar den virtuella U-disken och nätverket',
        tip2: 'I Endast-HID-läge är avbildningsmontering inaktiverat',
        rebuild: 'Byte av läge bygger upp USB-anslutningen på nytt. IronKVM startas inte om',
        enable: 'Aktivera Endast-HID-läge',
        disable: 'Inaktivera Endast-HID-läge'
      },
      resetHidDone: 'USB HID återställt',
      resetHidFailed: 'Återställning av USB HID misslyckades'
    },
    image: {
      driveLoaded: 'avbild isatt',
      driveWarning: 'se varningarna',
      warning: {
        missing: 'Avbildsfilen har raderats. Värden läser den gamla kopian tills du matar ut den.',
        writable: 'Läs och skriv: värden kan ändra den här avbilden.',
        tooBigForCd: 'För stor för cd-enheten ({{size}}, gräns {{max}}). Använd disken.',
        tooSmallForCd: 'För liten för cd-enheten ({{size}}). Använd disken.',
        empty: 'Filen är tom, troligen efter en misslyckad uppladdning eller nedladdning.'
      },
      delete: 'Ta bort',
      inUse: 'Används. Mata ut den innan du tar bort den.',
      retry: 'Försök igen',
      loadFailed: 'Det gick inte att läsa in listan över avbilder',
      readOnlyLocked: 'Mata ut disken för att ändra detta. Det gäller när en avbild sätts in.',
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
      deleteFailed: 'Borttagningen misslyckades',
      ventoy: {
        statusNoKernel: 'Stöds inte av denna firmware',
        statusNotInstalled: 'Inte installerad',
        statusReady: 'Redo',
        statusSelected: 'Valda avbilder: {{count}}',
        statusInDrive: 'I diskenheten, {{size}}',
        noKernel:
          'Kärnan i denna firmware saknar stöd för device-mapper, så Ventoy kan inte användas förrän en avbild med sådant stöd har installerats.',
        installDesc: 'Starta värden från flera avbilder på en disk, utan att kopiera dem.',
        install: 'Installera',
        installing: 'Laddar ner Ventoy, cirka 20 MB. Det kan ta några minuter.',
        needsData: 'Ventoy kräver en IronKVM-avbild med /data-partitionen monterad.',
        uninstall: 'Avinstallera',
        uninstallConfirm: 'Ta bort Ventoy-filerna?',
        noImages: 'Inga avbilder att lägga på Ventoy-disken.',
        onDisk: 'På Ventoy-disken',
        missing: 'Saknas: {{file}}',
        remove: 'Ta bort från Ventoy-disken',
        setHint:
          'Urvalet av avbilder kan bara ändras medan Ventoy-disken inte sitter i någon enhet.',
        useAsDisk: 'Använd som virtuell disk',
        failed: 'Ventoy-begäran misslyckades',
        secureBoot:
          'Med Secure Boot påslaget måste värden registrera Ventoys nyckel i MokManager en gång. Nyckelfilen ENROLL_THIS_KEY_IN_MOKMANAGER.cer finns på VTOYEFI-partitionen.',
        readOnly:
          'Värden ser disken som skrivskyddad, så Ventoy-persistens och ventoy.json på enheten fungerar inte.'
      },
      tips: {
        title: 'Hur man laddar upp',
        usb1: 'Anslut IronKVM till din dator via USB.',
        usb2: 'Säkerställ att den virtuella disken är monterad (Inställningar - Virtuell Disk).',
        usb3: 'Öppna den virtuella disken på din dator och kopiera avbildningsfilen till rotkatalogen.',
        scp1: 'Säkerställ att IronKVM och din dator är på samma lokala nätverk.',
        scp2: 'Öppna en terminal på din dator och använd SCP-kommandot för att ladda upp avbildningen till /data på IronKVM.',
        scp3: 'Exempel: scp din-avbildningssökväg root@din-nanokvm-ip:/data',
        tfCard: 'TF-kort',
        tf1: 'Denna metod stöds på Linux-system',
        tf2: 'Ta ut TF-kortet från IronKVM (för FULL-versionen, öppna chassit först).',
        tf3: 'Sätt in TF-kortet i en kortläsare och anslut till din dator.',
        tf4: 'Kopiera avbildningsfilen till /data på TF-kortet.',
        tf5: 'Sätt in TF-kortet i IronKVM.'
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
      close: 'Stäng',
      empty: 'Inga skript ännu. Ladda upp en .sh- eller .py-fil för att köra den på kortet.',
      loadFailed: 'Det gick inte att läsa in skripten',
      uploaded: 'Skript uppladdat',
      uploadFailed: 'Det gick inte att ladda upp skriptet',
      started: 'Skriptet startades i bakgrunden',
      deleteFailed: 'Det gick inte att ta bort skriptet',
      waitLimit: 'Väntar på att skriptet blir klart, i upp till {{minutes}} minuter.',
      timedOut:
        'Skriptet körde längre än {{minutes}} minuter och sidan slutade vänta. Det kan fortfarande köras på kortet.'
    },
    terminal: {
      invalidBaud: 'Den här baudhastigheten stöds inte.',
      invalidPort: 'Ange en enhetssökväg under /dev, till exempel /dev/ttyS1.',
      invalidSettings: 'Ogiltiga inställningar för serieporten. Detta är kortets eget skal.',
      disconnected: 'Frånkopplad. Tryck på Enter för att ansluta igen.',
      title: 'Terminal',
      nanokvm: 'IronKVM Terminal',
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
      no: 'Nej',
      yes: 'Ja',
      deleteConfirm: 'Ta bort den här sparade adressen?',
      delete: 'Ta bort',
      wake: 'Väck',
      rename: 'Byt namn',
      showMac: 'Visa MAC-adress',
      showName: 'Visa namn',
      requestFailed: 'Kunde inte nå enheten för att skicka kommandot',
      deleteFailed: 'Det gick inte att ta bort',
      renameFailed: 'Det gick inte att byta namn',
      title: 'Wake-on-LAN',
      sending: 'Skickar kommando...',
      sent: 'Kommando skickat',
      input: 'Ange MAC-adress',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'Uppladdningen misslyckades',
      uploadSuccess: 'Uppladdningen är klar',
      uploading: 'Laddar upp: {{file}}',
      downloadingPercent: 'Hämtar ({{percent}}): {{file}}',
      downloading: 'Hämtar: {{file}}',
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
      bootMenuPresent: '{{file}} finns redan på enheten med rätt kontrollsumma',
      bootMenuDesc:
        'Ladda ner netboot.xyz-ISO:n, med kontrollerad kontrollsumma, till den virtuella cd:n'
    },
    alerts: {
      title: 'Behöver åtgärd',
      temperature: {
        warning: 'Kortet håller {{celsius}} °C. Kontrollera att det får luft.',
        critical:
          'Kortet håller {{celsius}} °C, vilket är för varmt. Ge det luft eller stäng av det.'
      },
      storage: {
        warning:
          'Bara {{available}} av {{total}} ledigt på {{path}}. Stora avbilder kanske inte får plats.',
        critical:
          'Bara {{available}} ledigt på {{path}}. Uppladdningar, nedladdningar och installation av tillägg kommer att misslyckas. Radera avbilder du inte behöver.'
      },
      vpn: '{{name}} ska starta vid uppstart men körs inte, så fjärråtkomst via den fungerar inte.',
      openVpn: 'Öppna VPN-inställningar',
      stream:
        'Videoströmmen har misslyckats. Prova ett annat videoläge i menyn Skärm eller ladda om sidan.'
    },
    power: {
      resetDesc: 'Startar om värden direkt. Osparat arbete går förlorat.',
      powerShortDesc: 'Startar värden, eller ber dess OS att stänga av (ACPI).',
      powerLongDesc: 'Tvingar av värden utan avstängning.',
      hddLed: 'Disk-LED',
      hddActive: 'Aktiv',
      hddIdle: 'Inaktiv',
      title: 'Ström',
      showConfirm: 'Bekräftelse',
      showConfirmTip:
        'Fråga före ett kort tryck på strömknappen. Återställning och långt tryck frågar alltid.',
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
      ledConnectedFailed: 'Det gick inte att spara inställningen för ström-LED',
      powerLongConfirm: 'Hålla strömknappen i {{seconds}} s? Det bryter strömmen utan avstängning.',
      done: 'Knappen är tryckt',
      failed: 'Knapptrycket misslyckades'
    },
    settings: {
      title: 'Inställningar',
      nav: {
        system: 'System',
        network: 'Nätverk',
        access: 'Åtkomst',
        integrations: 'Integrationer',
        boot: 'Uppstart',
        browser: 'Den här webbläsaren',
        search: 'Sök en inställning',
        noMatch: 'Inga inställningar matchar',
        locked:
          'En åtgärd pågår. Andra sidor och stängning är inte tillgängliga förrän den är klar.',
        vpnProvider: 'VPN-leverantör'
      },
      mcp: {
        keyNote:
          'MCP använder en egen API-nyckel, som visas nedan. Nycklar från sidan API-nycklar fungerar inte här.',
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
        cancelBtn: 'Avbryt',
        showKey: 'Visa nyckel',
        hideKey: 'Dölj nyckel',
        regenerateKey: 'Skapa ny nyckel'
      },
      redfish: {
        example: 'Exempel',
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
      ipmi: {
        copyBeforeSave: 'Kopiera lösenordet nu. När det är sparat kan det inte visas igen.',
        noLogin:
          'IPMI är på, men inget aktivt konto har ett IPMI-lösenord, så ingen kan logga in. Ange ett nedan.',
        title: 'IPMI',
        warning:
          'IPMI-autentisering är svag till sin konstruktion. Den som når kortet och känner till ett användarnamn kan hämta en hash av användarens IPMI-lösenord och försöka knäcka den offline. Använd genererade lösenord, slå bara på IPMI i ett betrott nätverk och välj hellre Redfish över HTTPS där verktyget stöder det.',
        service: 'IPMI över LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) på UDP-port 623, för värdens ström och status. IPMI 1.5 och cipher suite 0 avvisas. När det stängs av avslutas alla IPMI-sessioner.',
        example: 'Exempel',
        copyFailed: 'Kopieringen misslyckades. Kopiera manuellt.',
        ledOn: 'Strömstatus, on, off, soft, cycle och reset är tillgängliga.',
        ledOff:
          '"Ström-LED ansluten" är av i strömmenyn, så strömtillståndet är okänt. Bara "power reset" fungerar: status, on, off, soft och cycle avvisas.',
        accounts: 'Konton',
        accountsDesc:
          'IPMI loggar in med KVM-kontona, vart och ett med sitt eget IPMI-lösenord, skilt från webblösenordet. Administratörer får ADMINISTRATOR. Användare får USER: de kan läsa strömtillståndet med "-L USER" men inte ändra det.',
        passwordSet: 'IPMI-lösenord angivet',
        passwordNotSet: 'Inget IPMI-lösenord: kan inte logga in över IPMI',
        nameTooLong: 'Namnet är längre än 16 tecken, vilket IPMI inte tillåter',
        accountDisabled: 'Kontot är inaktiverat',
        setPassword: 'Ange lösenord',
        changePassword: 'Byt lösenord',
        remove: 'Ta bort',
        removeConfirmTitle: 'Ta bort IPMI-lösenordet för {{user}}?',
        removeConfirmDesc:
          'Kontot kan inte längre logga in över IPMI, och dess IPMI-sessioner avslutas.',
        passwordTitle: 'IPMI-lösenord för {{user}}',
        passwordDesc:
          '12 till 20 skrivbara ASCII-tecken, skilt från webblösenordet. IPMI kräver att kortet sparar lösenordet i en form som det kan läsa tillbaka, så använd ett som inte används någon annanstans. Kopiera det innan du sparar: det visas inte igen.',
        passwordPlaceholder: 'IPMI-lösenord',
        generate: 'Generera',
        copy: 'Kopiera',
        save: 'Spara',
        passwordLength: 'Använd 12 till 20 tecken.',
        passwordChars: 'Använd bara skrivbara ASCII-tecken.',
        saved: 'IPMI-lösenordet sparat',
        failed: 'IPMI-åtgärden misslyckades',
        okBtn: 'Bekräfta',
        cancelBtn: 'Avbryt'
      },
      ssh: {
        service: 'SSH-server',
        serviceDesc: 'Starta sshd nu och vid varje uppstart',
        failed: 'Det gick inte att läsa in SSH-inställningarna',
        rootDefault: 'root har fortfarande fabrikslösenordet',
        rootEmpty: 'root har inget lösenord',
        rootWarning:
          'Alla som når konsolen eller SSH kan logga in som root. Ange ett lösenord under {{account}} > {{password}}: för enhetens ägare sätter det även roots lösenord.',
        connection: 'Anslutning',
        command: 'Logga in som root',
        port: 'Port',
        viaVpn: 'Via {{name}}',
        notRunning: 'sshd körs inte. Slå på SSH-servern för att ansluta.',
        hostKeys: 'Fingeravtryck för värdnycklar',
        hostKeysDesc: 'Jämför dem med det ssh visar vid första anslutningen.',
        noHostKeys: 'Inga värdnycklar ännu. sshd skapar dem första gången den startar.',
        keys: 'Auktoriserade nycklar',
        keysDesc:
          'Publika nycklar som kan logga in som root. De sparas på datapartitionen, så uppdateringar behåller dem.',
        noKeys: 'Inga auktoriserade nycklar ännu.',
        noComment: 'ingen kommentar',
        addPlaceholder:
          'Klistra in en publik nyckel, till exempel innehållet i ~/.ssh/id_ed25519.pub',
        add: 'Lägg till nyckel',
        added: 'Nyckel tillagd',
        removed: 'Nyckel borttagen',
        deleteConfirm: 'Ta bort den här nyckeln?',
        deleteConfirmDesc: 'Den kan inte längre logga in. Öppna sessioner förblir öppna.',
        invalidKey: 'Det här är ingen publik nyckel. Klistra in en enda rad från en .pub-fil.',
        keyOptions: 'Nycklar med alternativ som command= eller from= godtas inte här.',
        duplicateKey: 'Den här nyckeln är redan auktoriserad.',
        lastKey: 'Den sista nyckeln kan inte tas bort medan inloggning endast med nycklar är på.',
        keysOnly: 'Endast nycklar',
        keysOnlyDesc:
          'Stäng av inloggning med lösenord och keyboard-interactive. Öppna sessioner förblir öppna.',
        keysOnlyNeedsKey: 'Lägg först till en auktoriserad nyckel, annars kan ingen logga in.',
        keysOnlyOn: 'Inloggning med lösenord avstängd',
        keysOnlyOff: 'Inloggning med lösenord påslagen',
        notHonoured:
          'sshd i den här avbilden läser inte inställningen, så inloggning med lösenord förblir på.',
        reloadFailed:
          'Sparat, men sshd kunde inte läsas in på nytt. Det gäller nästa gång sshd startar.',
        notApplied:
          'sshd godtar fortfarande lösenord. Stäng av och slå på SSH-servern för att tillämpa inställningen.'
      },
      vnc: {
        address: 'Adress',
        certHint:
          'VeNCrypt X509Plain använder enhetens självsignerade certifikat, så klienten varnar vid första anslutningen. Godkänn det, eller spara certifikatet från den här sidans HTTPS-adress och ge det till TigerVNC med -X509CA=<fil>.',
        title: 'VNC',
        service: 'VNC-server',
        serviceDesc:
          'Låter en VNC-klient, till exempel TigerVNC eller Remmina, visa och styra värden. Klienten måste stödja Tight-kodning. En session i taget.',
        credentials:
          'Logga in med ett KVM-konto. Anslutningen krypteras med kortets TLS-certifikat (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'TCP-porten som servern lyssnar på.',
        maxFps: 'Gräns för bildfrekvens',
        maxFpsDesc: 'Det högsta antalet bilder per sekund som en klient får.',
        vncAuth: 'Enkel VNC-autentisering',
        vncAuthDesc:
          'För klienter utan VeNCrypt. Den kontrollerar ett separat VNC-lösenord i stället för ett konto.',
        vncAuthWarning:
          'Enkel VNC-autentisering krypterar inte anslutningen. Alla på nätverksvägen kan se skärmen och tangenttryckningarna. Använd den bara i ett betrott nätverk.',
        password: 'VNC-lösenord',
        passwordSet: 'Ett lösenord är angivet. Skriv ett nytt för att ändra det.',
        passwordInvalid: 'VNC-lösenordet måste vara 6 till 8 tecken.',
        save: 'Spara',
        saved: 'Inställningarna sparades',
        state: 'Status',
        listening: 'Lyssnar på port {{port}}',
        notListening: 'Lyssnar inte',
        noSession: 'Ingen öppen session',
        client: 'Klient',
        user: 'Användare',
        method: 'Autentisering',
        methodVencrypt: 'Konto över TLS',
        methodVnc: 'VNC-lösenord',
        since: 'Ansluten sedan',
        resolution: 'Upplösning',
        framesSent: 'Skickade bilder',
        lastError: 'Den senaste sessionen avslutades: {{error}}',
        refresh: 'Uppdatera',
        disconnect: 'Koppla från',
        disconnectConfirmTitle: 'Avsluta VNC-sessionen?',
        disconnectConfirmDesc:
          'Klienten kopplas från direkt, och alla tangenter och knappar som den håller ned släpps.',
        failed: 'VNC-åtgärden misslyckades',
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
        actionReset: 'Starta om',
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
        failed: 'Watchdog-åtgärden misslyckades',
        powerNeedsLed: 'Strömcykel kräver "Ström-LED ansluten" i strömmenyn.',
        noLedConfirmTitle: 'Slå på watchdog utan ström-LED?',
        noLedConfirmDesc:
          'Kortet kan inte se när värden är avstängd, så det behandlar värden som alltid på. Om du stänger av värden trycker watchdog på återställ när tidsgränsen har passerat. Anslut ström-LED:en för att undvika detta.',
        noLedConfirmOk: 'Slå på',
        cancel: 'Avbryt'
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
        title: 'Om IronKVM',
        information: 'Information',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Applikationsversion',
        applicationTip: 'IronKVM webbapplikationsversion',
        image: 'Systemversion',
        imageTip: 'IronKVM-kortavbild och NanoKVM-systemavbilden den bygger på',
        kernel: 'Kärnversion',
        kernelTip: 'Version av Linux-kärnan som körs nu',
        deviceKey: 'Enhetsnyckel',
        videoMemory: 'Videominne',
        videoMemoryTip:
          'Minne reserverat för videoinspelning. Det delas inte med resten av systemet.',
        videoMemoryGenerations_one: '{{count}} tidigare IronKVM-session håller videominne',
        videoMemoryGenerations_other: '{{count}} tidigare IronKVM-sessioner håller videominne',
        videoMemoryReboot: 'Starta om för att frigöra det.',
        community: 'Gemenskap',
        hostname: 'Värdnamn',
        hostnameUpdated: 'Värdnamn uppdaterat. Starta om för att tillämpa.',
        ipType: {
          Wired: 'Trådbundet',
          Wireless: 'Trådlöst',
          Other: 'Annat'
        },
        hostnameInvalid:
          'Använd bokstäver, siffror och bindestreck, högst 63 per punktseparerad del. Inget bindestreck i början eller slutet av en del.',
        hostnameFailed: 'Det gick inte att ändra värdnamnet',
        editHostname: 'Redigera värdnamn',
        docs: 'Dokumentation',
        hardware: 'Hårdvara',
        hardwareFaq: 'Vanliga frågor om hårdvaran',
        disclaimer:
          'IronKVM: härdad gemenskapsfirmware för Sipeed NanoKVM. Inte knuten till Sipeed.',
        basedOn: 'baserad på NanoKVM {{version}}'
      },
      preferences: {
        title: 'Preferenser'
      },
      performance: {
        title: 'Prestanda'
      },
      appearance: {
        thisBrowser: 'Den här webbläsaren',
        thisBrowserDesc:
          'Sparas bara i den här webbläsaren. Andra webbläsare har egna inställningar.',
        deviceWide: 'Enhet',
        deviceWideDesc: 'Sparas på enheten. Gäller alla som öppnar den.',
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
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Frontpanel'
        },
        hidModeDesc:
          'Prova endast HID-läge om värden inte tar emot tangentbord och mus. Det stänger av de virtuella enheterna och nätverket.',
        resetHidDesc:
          'Ansluter tangentbord och mus till värden igen. Använd om inmatningen slutar fungera.',
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
        hidOnly: 'Endast-HID-läge',
        hidOnlyDesc: 'Sluta emulera virtuella enheter, behåll bara grundläggande HID kontroll',
        disk: 'Virtuell disk',
        diskDesc: 'Montera virtuell U-disk på fjärrvärden',
        network: 'Virtuellt nätverk',
        networkDesc: 'Montera virtuell nätverkskort på fjärrvärden',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Värd:',
          description:
            'En privat nätverkslänk till fjärrvärden via USB-kabeln. Värden får en adress utan gateway och utan DNS, så den kan inte nå ditt lokala nätverk genom IronKVM.',
          off: 'Av',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (för värdar utan NCM)',
          rndis: 'RNDIS (erbjuds inte längre)',
          rndisNote: 'Den här länken använder RNDIS, som inte längre erbjuds. Välj NCM eller ECM.',
          subnet: 'Delnät',
          subnetDesc:
            'Ett privat IPv4-nätverk, /24 till /30. IronKVM tar den första adressen, värden den andra.',
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
          'Visa en seriell USB-port för fjärrvärden, för att logga in på denna IronKVM när nätverket inte går att nå',
        consoleTip:
          'Alla som styr fjärrvärden får en inloggningsprompt till denna IronKVM. Ställ in ett starkt lösenord innan du aktiverar (Konto - Byt lösenord).',
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
        rebootDesc: 'Är du säker på att du vill starta om IronKVM?',
        okBtn: 'Ja',
        cancelBtn: 'Nej',
        rebootFailed: 'Omstarten misslyckades'
      },
      network: {
        title: 'Nätverk',
        wifi: {
          disconnectBtn: 'Koppla från',
          disconnectWarning:
            'Om du når IronKVM via det här Wi-Fi-nätverket tappar sidan anslutningen.',
          disconnected: 'Wi-Fi frånkopplat',
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
          waitingHttp: 'Byter tillbaka till http. Ladda om sidan om den inte öppnas av sig själv.',
          failed: 'Det gick inte att ändra HTTPS-inställningen',
          enableConfirm: 'Slå på HTTPS?',
          disableConfirm: 'Stäng av HTTPS?',
          confirmDesc:
            'Detta loggar ut dig och startar om enhetens server, vilket tar ungefär två minuter. Sidan öppnar sedan {{url}}.',
          confirmOk: 'Fortsätt',
          confirmCancel: 'Avbryt'
        },
        ethernet: {
          title: 'IP-adress',
          description: 'Ställ in hur IronKVM får sin adress i det trådbundna nätverket',
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
          applyTitle: 'Vill du ändra adressen för IronKVM?',
          applyWarning:
            'Anslutningen till den här sidan går förlorad. IronKVM tillämpar den nya adressen och väntar {{seconds}} sekunder på att du når den där. Att nå den behåller ändringen. Om ingenting når den återställer IronKVM de tidigare inställningarna.',
          applyConfirm: 'Tillämpa',
          applyCancel: 'Avbryt',
          applyFailed: 'Adressen kunde inte tillämpas',
          trialTitle: 'Väntar på bekräftelse',
          trialDhcp: 'IronKVM begär en adress via DHCP.',
          trialStatic: 'IronKVM finns nu på {{address}}.',
          trialInstruction:
            'Öppna IronKVM på dess nya adress och logga in om den ber om det. Att nå den där behåller ändringen. Om ingenting når IronKVM inom {{seconds}} sekunder återställs de tidigare inställningarna.',
          trialOpen: 'Öppna den nya adressen',
          trialKeep: 'Behåll dessa inställningar',
          trialKept: 'Den nya adressen är sparad',
          trialKeepFailed: 'Inställningarna kunde inte behållas',
          trialGone: 'Ändringen har redan återställts. Försök igen.',
          unsaved: 'Osparade ändringar'
        },
        dns: {
          title: 'DNS',
          description: 'Konfigurera DNS-servrar för IronKVM',
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
        connect: 'Anslut',
        connectDesc: 'Gå med i {{name}}-nätverket. Av kopplar från utan att stoppa tjänsten.',
        kvmUrl: 'KVM-adress',
        moreTip: 'Fler åtgärder',
        restartTip: 'Starta om',
        stopTip: 'Stoppa',
        updateTip: 'Uppdatera till {{version}}',
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
          tip: 'Om tjänsten får ont om minne kan du prova att slå på swap. Det ställs in under "Inställningar > Prestanda".'
        },
        copy: 'Kopiera',
        copied: 'Länken kopierad',
        copyFailed: 'Det gick inte att kopiera länken. Markera den och kopiera den för hand.',
        open: 'Öppna',
        checkAgain: 'Kontrollera igen',
        notSignedIn: 'Inte inloggad än. Slutför inloggningen via länken och kontrollera igen.',
        checkFailed: 'Det gick inte att kontrollera inloggningsstatus',
        loginWaiting:
          'Sidan kontrollerar med några sekunders mellanrum och fortsätter när du har loggat in.',
        uninstallFailed: 'Avinstallationen misslyckades',
        loginFailed: 'Inloggningen misslyckades'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Ladda ner',
        package: 'installationspaketet',
        unzip: 'och packa upp det',
        notLogin: 'Enheten är ännu inte bunden. Logga in och bind enheten till ditt konto.',
        urlPeriod: 'Denna URL är giltig i 10 minuter',
        login: 'Logga in',
        logout: 'Logga ut',
        logoutDesc: 'Är du säker på att du vill logga ut?',
        manualIntro: 'Eller installera det manuellt över SSH:',
        copyBinaries: 'Kopiera tailscale och tailscaled till {{dir}} på IronKVM',
        linksFile: 'Skapa i samma katalog en fil som heter links med de här två raderna:',
        rebootRefresh: 'Starta om IronKVM och uppdatera sedan den här sidan'
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
        logout: 'Avregistrera',
        logoutDesc:
          'Avregistrering tar bort den här noden från ditt NetBird-konto och raderar dess konfiguration här. För att ansluta igen krävs en installationsnyckel eller en SSO-inloggning, och noden kan få en ny IP. Fortsätta?',
        joinFailed: 'Kunde inte ansluta till nätverket'
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
            'SHA-512 kontrollerar endast att paketet överensstämmer med manifestet från den här servern. Det bevisar inte att paketet är en officiell IronKVM-utgåva. En felaktig eller skadlig server kan göra enheten obrukbar, orsaka dataförlust eller äventyra systemets säkerhet.',
          confirm: 'Använd ändå',
          useSipeed: 'Använd Sipeeds officiella server',
          previewDisabled:
            'Förhandsuppdateringar är inte tillgängliga när en anpassad uppdateringsserver är aktiverad.'
        },
        offline: {
          chooseFile: 'Välj fil',
          installing: 'Uppladdningen är klar. Installerar...',
          noFile: 'Ingen fil vald',
          title: 'Offlineuppdateringar',
          desc: 'Uppdatera genom lokalt installationspaket',
          upload: 'Ladda upp',
          checksumPlaceholder: 'SHA-256-kontrollsumma (valfri)',
          invalidChecksum: 'SHA-256-kontrollsumman måste innehålla 64 hexadecimala tecken.',
          checksumMismatch: 'SHA-256-verifieringen misslyckades. Paketet kan vara skadat.',
          invalidName: 'Ogiltigt filnamnsformat. Ladda ner från GitHub-versioner.',
          updateFailed: 'Uppdatering misslyckades. Försök igen.'
        },
        updateTo: 'Uppdatera till {{version}}',
        updateConfirmDesc:
          'Enheten installerar uppdateringen och startar om sin server. Sidan laddas om när servern är tillbaka.',
        releaseNotes: 'Versionsinformation'
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
        mcpNote: 'Dessa nycklar fungerar inte för MCP, som har en egen nyckel på MCP-sidan.',
        metricsUrl: 'Mätvärdes-URL',
        monitoring: 'Övervakning',
        monitoringDesc:
          'Prometheus läser mätvärdena med en API-nyckel från den här sidan, skickad som Bearer-token. Alla roller kan läsa dem.',
        scrapeConfig: 'Prometheus scrape-konfiguration',
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
        kvmDescription: 'Manövrera fjärrvärden genom IronKVM.',
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
          restoring: 'Återställer PicoClaw',
          ready: 'Runtime klar',
          stopped: 'Runtime stoppad',
          blockedByMCP: 'Extern MCP-styrning är aktiv',
          readyBlockedByMCP: 'Runtime körs, men extern MCP styr just nu enhetens inmatning.',
          readyWithoutControl: 'Runtime körs. Ge PicoClaw enhetsstyrning innan du ansluter igen.',
          unavailable: 'Runtime inte tillgänglig',
          configError: 'Konfigurationsfel'
        },
        transport: {
          connecting: 'Ansluter',
          connected: 'Ansluten',
          disconnected: 'Frånkopplad',
          reconnect: 'Anslut igen',
          reconnectDescription: 'Anslut igen till den pågående PicoClaw-sessionen.',
          reconnectBlocked: 'PicoClaw behöver enhetsstyrning innan den kan ansluta igen.'
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
        picoclawDescription:
          'PicoClaw kan skicka tangentbords- och musinmatning. Manuell inmatning kan pausas.',
        mcp: 'Enhetsstyrning: extern MCP',
        mcpDescription: 'Extern MCP kan skriva till enheten. PicoClaw tar inte över inmatningen.',
        off: 'Enhetsstyrning: av',
        offDescription:
          'AI skickar ingen tangentbords- eller musinmatning. Manuell styrning finns kvar.',
        transitioning: 'Enhetsstyrning: byter',
        transitioningDescription: 'Enhetsstyrningen synkroniseras. Vänta.',
        grant: 'Ge styrning',
        release: 'Släpp',
        releasing: 'Släpper...',
        switching: 'Byter...',
        releasingLabel: 'Enhetsstyrning: släpper',
        releasingDescription:
          'Enhetsstyrningen lämnas tillbaka. PicoClaw har stoppat pågående inmatning.',
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
        switchFromMCP: 'Byt till PicoClaw och starta',
        takeoverAndStart: 'Ta över och starta'
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
    upstream: {
      check: 'Sök efter uppdateringar',
      updateTo: 'Uppdatera till {{version}}',
      confirm: 'Uppdatera {{name}} till {{version}}?',
      confirmDesc:
        'Den nya utgåvan hämtas från GitHub och kontrolleras mot de kontrollsummor den publicerar. Om något misslyckas behålls den nuvarande versionen.',
      ok: 'Uppdatera',
      upToDate: 'Uppdaterad',
      builtIn: 'inbyggd',
      checkFailed: 'Kunde inte söka efter uppdateringar: {{error}}',
      unverifiable: 'Version {{version}} erbjuds inte: {{reason}}',
      inUse: 'Kan inte uppdatera nu: {{reason}}',
      running: 'Uppdaterar till {{version}}...',
      done: '{{name}} uppdaterad till {{version}}',
      failed: 'Den senaste uppdateringen misslyckades: {{error}}'
    },
    menu: {
      mediaNetboot: 'Inställningar för nätverksstart',
      collapse: 'Fäll ihop menyn',
      expand: 'Expandera menyn',
      more: 'Mer',
      media: 'Media',
      tools: 'Verktyg',
      text: 'Text',
      advanced: 'Avancerat',
      mediaMounted: 'Monterad',
      mediaLibrary: 'Bibliotek',
      mediaBoot: 'Uppstart',
      textToHost: 'Till värden',
      textFromHost: 'Från värden'
    },
    ion: {
      checking: 'Kontrollerar videominnet innan strömmen startar...',
      warn: 'Lite videominne kvar. En omstart av servern skulle ta slut på det. Starta om när det passar.',
      criticalTitle: 'Inte tillräckligt med videominne för att starta strömmen',
      criticalBody:
        'Att starta video skulle ta slut på det reserverade minnet och stoppa servern. Alla andra funktioner fungerar fortfarande, även strömstyrning och omstart. Endast en omstart av IronKVM frigör detta minne.',
      criticalContinue: 'Starta video ändå',
      criticalReboot: 'Starta om IronKVM',
      criticalRebooting: 'Startar om...'
    }
  }
};

export default se;
