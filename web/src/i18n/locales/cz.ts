const cz = {
  translation: {
    head: {
      desktop: 'Vzdálená plocha',
      login: 'Přihlášení',
      changePassword: 'Změna hesla',
      terminal: 'Terminál',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        'Prohlížeč odmítl uložit relaci. Cookie, které zůstalo z předchozí relace přes HTTPS, nelze přes nešifrované http nahradit. Vymažte cookies pro tuto adresu nebo otevřete anonymní okno a přihlaste se znovu.',
      login: 'Přihlášení',
      placeholderUsername: 'Zadejte prosím uživatelské jméno',
      placeholderPassword: 'Zadejte prosím heslo',
      placeholderCurrentPassword: 'Aktuální heslo',
      placeholderPassword2: 'Zadejte prosím heslo znovu',
      noEmptyUsername: 'Uživatelské jméno nesmí být prázdné',
      noEmptyPassword: 'Heslo nesmí být prázdné',
      passwordLength: 'Heslo musí mít 8 až 72 znaků',
      noAccount:
        'Nepodařilo se získat informace o uživateli, prosím obnovte stránku nebo resetujte heslo',
      invalidUser: 'Neplatné uživatelské jméno nebo heslo',
      locked: 'Příliš mnoho přihlášení, zkuste to znovu později',
      globalLocked: 'Systém je chráněn, zkuste to znovu později',
      error: 'Neočekávaná chyba',
      invalidCurrentPassword: 'Aktuální heslo je nesprávné',
      changePassword: 'Změnit heslo',
      changePasswordDesc:
        'Pro bezpečnost vašeho zařízení prosím změňte heslo pro přihlášení na webu.',
      differentPassword: 'Hesla se neshodují',
      illegalUsername: 'Uživatelské jméno obsahuje nepovolené znaky',
      illegalPassword: 'Heslo obsahuje nepovolené znaky',
      forgetPassword: 'Zapomenuté heslo',
      ok: 'OK',
      cancel: 'Zrušit',
      loginButtonText: 'Přihlášení',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the NanoKVM for 10 seconds.',
        reset2: 'Podrobné kroky najdete v tomto dokumentu:',
        reset3: 'Výchozí webový účet:',
        reset4: 'Výchozí účet SSH:',
        change1: 'Upozorňujeme, že tato akce změní následující hesla:',
        change2: 'Heslo pro webové přihlášení',
        change3: 'Heslo systémového uživatele root (heslo pro přihlášení SSH)',
        change4: 'Chcete-li hesla resetovat, podržte tlačítko BOOT na NanoKVM.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Nastavit Wi-Fi pro NanoKVM',
      success: 'Please check the network status of NanoKVM and visit the new IP address.',
      failed: 'Operace selhala, zkuste to znovu.',
      invalidMode:
        'Aktuální režim nepodporuje nastavení sítě. Přejděte do svého zařízení a povolte konfigurační režim Wi-Fi.',
      confirmBtn: 'Ok',
      finishBtn: 'Dokončeno',
      ap: {
        authTitle: 'Vyžaduje se ověření',
        authDescription: 'Pokračujte zadáním hesla AP',
        authFailed: 'Neplatné heslo AP',
        passPlaceholder: 'AP heslo',
        verifyBtn: 'Ověřte'
      }
    },
    screen: {
      scale: 'Měřítko',
      title: 'Obrazovka',
      video: 'Režim videa',
      videoDirectTips: 'Chcete-li používat tento režim, povolte HTTPS v "Nastavení > Zařízení"',
      resolution: 'Rozlišení',
      controlRegion: {
        title: 'Kalibrace myši',
        description:
          'Toto nastavení použijte, pokud ovládané zařízení používá jiné rozlišení než 16:9 a kurzor je vodorovně nebo svisle posunutý.',
        off: 'Vypnuto',
        auto: 'Automaticky',
        autoWarning: 'Kalibrace může selhat, pokud má uživatelská aplikace zcela černé pozadí.',
        manual: 'Ručně',
        selectedResolution: 'Rozlišení vybrané oblasti',
        unused: 'Nepoužito',
        originalResolution: 'Původní rozlišení',
        selectResolution: 'Vyberte původní rozlišení',
        addResolution: 'Přidat vlastní rozlišení',
        add: 'Přidat',
        duplicateResolution: 'Toto rozlišení již existuje.',
        width: 'Šířka',
        height: 'Výška',
        apply: 'Vypočítat a použít',
        invalidResolution: 'Po načtení videa zadejte platné původní rozlišení.',
        select: 'Vybrat oblast',
        clear: 'Obnovit automatickou detekci',
        saveFailed: 'Vstupní oblast se nepodařilo uložit.',
        tooSmall: 'Vybraná oblast je příliš malá.',
        previewUnavailable: 'Náhled není k dispozici',
        clearConfirm: 'Obnovit automatickou detekci černých okrajů?',
        dragHint: 'Tažením vyberte oblast vzdálené plochy',
        finish: 'Hotovo',
        confirm: 'Potvrdit',
        cancel: 'Zrušit'
      },
      auto: 'Automatické',
      autoTips:
        'Může docházet k trhání obrazu nebo posunu myši při určitých rozlišeních. Zvažte úpravu rozlišení vzdáleného hostitele nebo vypněte automatický režim.',
      fps: 'FPS',
      customizeFps: 'Přizpůsobit',
      quality: 'Kvalita',
      qualityLossless: 'Bezeztrátový',
      qualityHigh: 'Vysoký',
      qualityMedium: 'Střední',
      qualityLow: 'Nízký',
      frameDetect: 'Detekce snímků',
      frameDetectTip:
        'Vypočítá rozdíl mezi snímky. Přenos video streamu se zastaví, pokud nejsou detekovány změny na obrazovce vzdáleného hostitele.',
      resetHdmi: 'Resetovat HDMI',
      mixedH264: {
        title: 'Konflikt streamu H.264',
        description:
          'H.264 Direct a H.264 WebRTC se používají současně. To může způsobit trhání obrazu nebo poškozené video. Používejte pouze jeden režim H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Připojení WebRTC se nezdařilo',
        description: 'Zkontrolujte síťové připojení nebo přepněte režim videa.'
      },
      captureStatus: {
        hdmiError: 'Chyba obrazu HDMI',
        unsupportedResolution: 'Aktuální rozlišení není podporováno',
        retrieving: 'Načítá se obraz...',
        changingResolution: 'Přepíná se rozlišení...',
        updateFailed: 'Obraz se teď nemůže aktualizovat',
        videoError: 'Chyba zobrazení videa',
        noHdmi: 'Nebyl zjištěn signál HDMI',
        unavailable: 'Obraz teď nelze zobrazit'
      }
    },
    keyboard: {
      title: 'Klávesnice',
      paste: 'Vložit',
      tips: 'Napíše text na hostiteli jako stisky kláves. Zvolte rozložení klávesnice, které hostitel používá.',
      placeholder: 'Zadejte text',
      submit: 'Odeslat',
      virtual: 'Klávesnice',
      readClipboard: 'Přečíst ze schránky',
      clipboardPermissionDenied:
        'Oprávnění ke schránce odepřeno. Povolte prosím přístup do schránky ve svém prohlížeči.',
      clipboardReadError: 'Nepodařilo se přečíst schránku',
      mediaKeys: {
        title: 'Multimediální klávesy',
        mute: 'Ztlumit',
        volumeDown: 'Snížit hlasitost',
        volumeUp: 'Zvýšit hlasitost',
        previous: 'Předchozí skladba',
        playPause: 'Přehrát nebo pozastavit',
        next: 'Další skladba',
        stop: 'Zastavit'
      },
      pasting: {
        layout: 'Rozložení klávesnice na hostiteli',
        layouts: {
          us: 'Angličtina (USA)',
          uk: 'Angličtina (Spojené království)',
          de: 'Němčina',
          fr: 'Francouzština',
          es: 'Španělština',
          it: 'Italština',
          ptBr: 'Portugalština (Brazílie)',
          se: 'Švédština / finština',
          ru: 'Ruština',
          ja: 'Japonština',
          ko: 'Korejština'
        },
        speed: 'Rychlost psaní',
        speeds: {
          fast: 'Rychlá',
          normal: 'Normální',
          slow: 'Pomalá'
        },
        estimate: 'Doba psaní: asi {{duration}}',
        untypeable: 'Znaky, které toto rozložení nenapíše: {{count}}',
        untypeableAt: 'řádek {{line}}, sloupec {{column}}',
        skipUntypeable: 'Napsat zbytek',
        shortcut: '{{shortcut}} napíše obsah schránky na hostiteli hned.',
        clipboardUnavailable:
          'Prohlížeč dovolí stránce číst schránku jen přes HTTPS. Vložte text do pole pomocí Ctrl+V.',
        clipboardEmpty: 'Schránka neobsahuje text.',
        tooLong: 'Text je příliš dlouhý. Limit je {{max}} znaků.',
        inProgress: 'Už se píše jiný vložený text.',
        typing: 'Píše se na hostiteli',
        done: 'Text napsán',
        canceled: 'Vkládání zrušeno',
        failed: 'Vkládání selhalo',
        cancel: 'Zrušit',
        controlBusy: 'Klávesnici používá jiný ovladač.',
        hidError: 'Stisky kláves se nepodařilo odeslat hostiteli.'
      },
      shortcut: {
        title: 'Zkratky',
        custom: 'Vlastní',
        capture: 'Kliknutím sem zachytíte zástupce',
        clear: 'Jasno',
        save: 'Uložit',
        captureTips:
          'Zachycení systémových kláves (například klávesy Windows) vyžaduje oprávnění pro celou obrazovku.',
        enterFullScreen: 'Přepnout režim celé obrazovky.'
      },
      leaderKey: {
        title: 'Klávesa Leader',
        desc: 'Obejít omezení prohlížeče a odeslat systémové zkratky přímo vzdálenému hostiteli.',
        howToUse: 'Jak používat',
        simultaneous: {
          title: 'Simultánní režim',
          desc1: 'Stiskněte a podržte klávesu Leader a poté stiskněte zkratku.',
          desc2: 'Intuitivní, ale může být v rozporu se systémovými zkratkami.'
        },
        sequential: {
          title: 'Sekvenční režim',
          desc1:
            'Stiskněte klávesu Leader → postupně stiskněte zkratku → znovu stiskněte klávesu Leader.',
          desc2: 'Vyžaduje více kroků, ale zcela se vyhne systémovým konfliktům.'
        },
        enable: 'Povolit klávesu Leader',
        tip: 'Když je tato klávesa nastavena jako klávesa Leader, slouží pouze jako spouštěč zkratek a ztrácí své výchozí chování.',
        placeholder: 'Stiskněte klávesu Leader',
        shiftRight: 'Pravý Shift',
        ctrlRight: 'Pravý Ctrl',
        metaRight: 'Pravý Win',
        submit: 'Odeslat',
        recorder: {
          rec: 'REC',
          activate: 'Aktivovat klávesy',
          input: 'Stiskněte prosím zkratku...'
        }
      }
    },
    mouse: {
      title: 'Myš',
      cursor: 'Styl kurzoru',
      default: 'Výchozí kurzor',
      pointer: 'Ukazovací kurzor',
      cell: 'Kurzor buňky',
      text: 'Textový kurzor',
      grab: 'Chytnout kurzor',
      hide: 'Skrýt kurzor',
      mode: 'Režim myši',
      absolute: 'Absolutní režim',
      relative: 'Relativní režim',
      absoluteShort: 'Absolutní',
      relativeShort: 'Relativní',
      absoluteStalled: 'Cílové zařízení ignoruje absolutní myš',
      absoluteStalledDesc:
        'Cílové zařízení přestalo přijímat hlášení absolutní myši, takže se pohyby kurzoru ztrácejí. Klávesnice není dotčena. Často pomůže obnovení USB; relativní režim používá jiný koncový bod.',
      useRelative: 'Přepnout na relativní režim',
      direction: 'Směr kolečka',
      scrollUp: 'Přejděte nahoru',
      scrollDown: 'Přejděte dolů',
      speed: 'Rychlost kolečka',
      fast: 'Rychle',
      slow: 'Pomalu',
      requestPointer:
        'Používá se relativní režim. Klikněte prosím na plochu pro získání kurzoru myši.',
      resetHid: 'Resetovat HID',
      hidOnly: {
        title: 'Režim pouze HID',
        desc: 'Pokud vaše myš a klávesnice přestanou reagovat a resetování HID nepomůže, může jít o problém s kompatibilitou mezi NanoKVM a zařízením. Zkuste povolit režim HID-Only pro lepší kompatibilitu.',
        tip1: 'Povolení režimu HID-Only odpojí virtuální U-disk a virtuální síť',
        tip2: 'V režimu HID-Only je připojení obrazu zakázáno',
        rebuild: 'Přepnutí režimu znovu sestaví připojení USB. NanoKVM se nerestartuje',
        enable: 'Povolit režim HID-Only',
        disable: 'Zakázat režim HID-Only'
      }
    },
    image: {
      title: 'Obrázky',
      loading: 'Načítání...',
      empty: 'Nic nenalezeno',
      mountMode: 'Režim připojení',
      mountFailed: 'Připojení se nezdařilo',
      mountDesc:
        'V některých systémech je nutné před připojením obrazu vysunout virtuální disk na vzdáleném hostiteli.',
      unmountFailed: 'Odpojení se nezdařilo',
      unmountDesc:
        'Na některých systémech se musíte před odpojením obrazu ručně vysunout ze vzdáleného hostitele.',
      refresh: 'Obnovte seznam obrázků',
      disk: 'Disk',
      cdrom: 'CD',
      driveEmpty: 'Prázdná',
      eject: 'Vysunout',
      readOnly: 'Jen pro čtení',
      readOnlyTip: 'Platí pro další obraz vložený do disku.',
      noDrives: 'Žádné virtuální jednotky. Zapněte virtuální disk v Nastavení.',
      insertFailed: 'Vložení se nezdařilo',
      ejectFailed: 'Vysunutí se nezdařilo',
      insertInto: 'Vložit do jednotky {{drive}}. Kliknutím změníte.',
      loadedIn: 'V jednotce {{drive}}',
      attention: 'Pozor',
      deleteConfirm: 'Opravdu chcete smazat tento obrázek?',
      okBtn: 'Ano',
      cancelBtn: 'Ne',
      tips: {
        title: 'Jak nahrát',
        usb1: 'Připojte NanoKVM k vašemu počítači přes USB.',
        usb2: 'Ujistěte se, že je virtuální disk připojen (Nastavení - Virtuální disk).',
        usb3: 'Otevřete virtuální disk na vašem počítači a zkopírujte soubor s obrazem do kořenového adresáře virtuálního disku.',
        scp1: 'Ujistěte se, že jsou NanoKVM a váš počítač ve stejné místní síti.',
        scp2: 'Otevřete terminál na vašem počítači a použijte příkaz SCP pro nahrání souboru s obrazem do adresáře /data na zařízení NanoKVM.',
        scp3: 'Příklad: scp cesta-k-vašemu-obrazu root@ip-nanokvm:/data',
        tfCard: 'SD Karta',
        tf1: 'Tato metoda je podporována na systémech Linux',
        tf2: 'Vyjměte SD kartu z NanoKVM (u plné verze nejprve rozložte krabičku).',
        tf3: 'Vložte SD kartu do čtečky karet a připojte ji k vašemu počítači.',
        tf4: 'Zkopírujte soubor s obrazem do adresáře /data na SD kartě.',
        tf5: 'Vložte SD kartu zpět do NanoKVM.'
      }
    },
    script: {
      title: 'Skript',
      upload: 'Nahrát',
      run: 'Spustit',
      runBackground: 'Spustit na pozadí',
      runFailed: 'Spuštění se nezdařilo',
      attention: 'Pozor',
      delDesc: 'Opravdu chcete tento soubor smazat?',
      confirm: 'Ano',
      cancel: 'Ne',
      delete: 'Smazat',
      close: 'Zavřít'
    },
    terminal: {
      title: 'Terminál',
      nanokvm: 'Terminál NanoKVM',
      serial: 'Terminál sériového portu',
      serialPort: 'Sériový port',
      serialPortPlaceholder: 'Zadejte prosím sériový port',
      baudrate: 'Přenosová rychlost',
      parity: 'Parita',
      parityNone: 'Žádné',
      parityEven: 'Sudá',
      parityOdd: 'Lichá',
      flowControl: 'Řízení toku',
      flowControlNone: 'Žádné',
      flowControlSoft: 'Softwarové',
      flowControlHard: 'Hardwarové',
      dataBits: 'Datové bity',
      stopBits: 'Stop bity',
      confirm: 'OK'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'Odesílání příkazu...',
      sent: 'Příkaz odeslán',
      input: 'Zadejte prosím MAC adresu',
      ok: 'OK'
    },
    download: {
      title: 'Stahovač obrazů',
      input: 'Zadejte prosím vzdálený obrázek URL',
      ok: 'OK',
      disabled: 'Oddíl /data je RO, takže obrázek nelze stáhnout',
      uploadbox: 'Přetáhněte soubor sem nebo kliknutím vyberte',
      inputfile: 'Zadejte soubor obrázku',
      NoISO: 'Žádné ISO',
      sha256: 'SHA-256 (volitelné)',
      sha256Placeholder: 'Zadejte 64znakový kontrolní součet SHA-256',
      invalidSHA256: 'SHA-256 musí být 64znakový hexadecimální řetězec',
      failed: 'Stažení se nezdařilo',
      success: 'Stažení proběhlo úspěšně',
      checksumFailed: 'Stažení se nezdařilo: ověření SHA-256 selhalo',
      cancel: 'Zrušit',
      cancelFailed: 'Stažení se nepodařilo zrušit'
    },
    power: {
      title: 'Napájení',
      showConfirm: 'Potvrzení',
      showConfirmTip: 'Výkonové operace vyžadují další potvrzení',
      reset: 'Resetovat',
      power: 'Napájení',
      powerShort: 'Napájení (krátký stisk)',
      powerLong: 'Napájení (dlouhý stisk)',
      resetConfirm: 'Pokračovat v operaci resetování?',
      powerConfirm: 'Pokračovat v napájení?',
      okBtn: 'Ano',
      cancelBtn: 'Ne',
      hostOs: 'OS hostitele',
      hostOsTip: 'Odesílá se jako klávesy USB. Co udělají, rozhoduje hostitel.',
      sleep: 'Uspat',
      wake: 'Probudit',
      wakeKey: 'Probudit klávesou Shift',
      powerDown: 'Vypnout',
      sleepConfirm: 'Uspat hostitele?',
      powerDownConfirm: 'Odeslat hostiteli klávesu vypnutí?',
      wakeTip:
        'Uspaný hostitel často ignoruje Probudit od zařízení, které ho uspalo. Probudit klávesou Shift stiskne klávesu na klávesnici, kterou přijme více hostitelů.',
      led: 'LED napájení',
      ledOn: 'Svítí',
      ledOff: 'Nesvítí',
      ledUnknown: 'Neznámý',
      ledConnected: 'LED napájení připojena',
      ledConnectedTip:
        'Zapněte, jen pokud je konektor LED napájení hostitele propojen s deskou. Bez něj je stav napájení neznámý.',
      ledConnectedFailed: 'Nastavení LED napájení se nepodařilo uložit'
    },
    settings: {
      title: 'Nastavení',
      mcp: {
        title: 'Služba MCP',
        service: 'Vzdálené ovládání MCP',
        serviceDesc:
          'Umožnit důvěryhodným klientům MCP ovládat klávesnici a myš a pořizovat snímky obrazovky',
        securityWarning:
          'Kdokoli s tímto API klíčem může ovládat vzdálený hostitel a zobrazit jeho obrazovku. Používejte HTTPS a povolte službu pouze v důvěryhodných sítích.',
        endpoint: 'Koncový bod',
        apiKey: 'API klíč',
        regenerateConfirmTitle: 'Vygenerovat nový MCP API klíč?',
        regenerateConfirmDesc: 'Aktuální klíč přestane okamžitě fungovat.',
        enableConfirmTitle: 'Povolit externí ovládání MCP?',
        enableConfirmDesc:
          'Povolením MCP se zastaví PicoClaw a ukončí se všechny aktivní relace PicoClaw.',
        failed: 'Operace MCP se nezdařila',
        copyFailed: 'Kopírování se nezdařilo. Zkopírujte ručně.',
        okBtn: 'Potvrdit',
        cancelBtn: 'Zrušit'
      },
      redfish: {
        title: 'Redfish',
        service: 'Služba Redfish',
        serviceDesc:
          'Rozhraní DMTF Redfish API pro ovládání napájení, virtuální média a stav z nástrojů jako redfishtool a Ansible. Vypnutím se ukončí všechny relace Redfish.',
        endpoint: 'Kořen služby',
        httpsOn: 'Deska poskytuje HTTPS, které většina nástrojů Redfish potřebuje.',
        httpsOff:
          'Deska poskytuje pouze nešifrované HTTP. Většina nástrojů Redfish potřebuje HTTPS: zapněte ho v "Nastavení > Síť".',
        credentials:
          'Redfish přijímá účty KVM s ověřením Basic nebo s relací Redfish a také klíče API zasílané jako X-Auth-Token. Klíče API se spravují na stránce Klíče API.',
        powerActions: 'Akce napájení',
        powerActionsDesc:
          'Typy resetu, které jsou nyní nabízeny. On, ForceOff a GracefulShutdown potřebují znát stav napájení, proto se nabízejí jen tehdy, když je v nabídce napájení zapnuto "LED napájení připojena".',
        sessions: 'Relace',
        noSessions: 'Žádné otevřené relace Redfish',
        created: 'Vytvořeno',
        lastUsed: 'Naposledy použito',
        refresh: 'Obnovit',
        end: 'Ukončit',
        endConfirmTitle: 'Ukončit tuto relaci Redfish?',
        endConfirmDesc:
          'Její token okamžitě přestane fungovat. Klient se bude muset znovu přihlásit.',
        failed: 'Operace Redfish se nezdařila',
        copyFailed: 'Kopírování se nezdařilo. Zkopírujte ručně.',
        okBtn: 'Potvrdit',
        cancelBtn: 'Zrušit'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Watchdog hostitele',
        serviceDesc:
          'Pokud má hostitel běžet a jeho obraz se po dobu časového limitu nezmění nebo chybí signál HDMI, deska stiskne reset nebo hostitele vypne a znovu zapne.',
        stillWarning:
          'Hostitel, jehož displej přejde do spánku nebo jehož obraz se při práci nemění, vypadá jako zamrzlý. Vypněte na hostiteli spánek displeje, nebo zadejte adresu pro ping.',
        ledHint:
          'Přepínač "LED napájení připojena" je v nabídce napájení vypnutý. Watchdog nevidí, kdy je hostitel vypnutý, a proto ho považuje za stále zapnutý.',
        timeout: 'Časový limit',
        timeoutDesc: 'Jak dlouho smí hostitel nejevit známky života, než watchdog zasáhne.',
        action: 'Akce',
        actionDesc: 'Vypnutí a zapnutí drží tlačítko napájení 5 sekund a pak ho stiskne znovu.',
        actionReset: 'Reset',
        actionPower: 'Vypnout a zapnout',
        cooldown: 'Prodleva',
        cooldownDesc: 'Nejkratší doba mezi dvěma akcemi.',
        maxPerHour: 'Akcí za hodinu',
        maxPerHourDesc: 'Nejvyšší počet akcí za hodinu.',
        pingHost: 'Adresa pro ping',
        pingHostDesc:
          'IP adresa hostitele. Odpověď se počítá jako známka života. Nechte prázdné, pokud nechcete pingovat.',
        pingHostInvalid: 'Zadejte adresu IPv4 nebo IPv6.',
        minutes: 'min',
        save: 'Uložit',
        saved: 'Uloženo',
        state: 'Detektor',
        status: {
          off: 'Vypnuto',
          watching: 'Sleduje',
          hostOff: 'Hostitel vypnut',
          captureOff: 'Snímání HDMI vypnuto',
          cooldown: 'Prodleva',
          capped: 'Dosažen hodinový limit',
          acting: 'Zasahuje'
        },
        signal: 'Signál HDMI',
        yes: 'Ano',
        no: 'Ne',
        led: 'LED napájení',
        on: 'Svítí',
        off: 'Nesvítí',
        ledNotConnected: 'Nepřipojeno',
        ping: 'Ping',
        pingNotSet: 'Nenastaveno',
        pingReply: 'Odpovídá',
        pingNoReply: 'Bez odpovědi',
        lastChange: 'Poslední změna obrazu',
        never: 'Nikdy',
        actsIn: 'Zasáhne za',
        actionsLastHour: 'Akce za poslední hodinu',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Záznam',
        noLog: 'Watchdog zatím nezasáhl.',
        refresh: 'Obnovit',
        reasonFrozen: 'Obraz se nezměnil',
        reasonNoSignal: 'Žádný signál HDMI',
        stuckFor: 'bez známek života po dobu {{duration}}',
        pressFailed: 'Stisk se nezdařil: {{error}}',
        noScreenshot: 'Bez snímku obrazovky',
        failed: 'Operace watchdogu selhala'
      },
      about: {
        title: 'O NanoKVM',
        information: 'Informace',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Verze aplikace',
        applicationTip: 'Verze webové aplikace NanoKVM',
        image: 'Verze obrazu',
        imageTip: 'Verze systémového obrazu NanoKVM',
        kernel: 'Verze jádra',
        kernelTip: 'Verze aktuálně běžícího jádra Linuxu',
        deviceKey: 'Klíč zařízení',
        videoMemory: 'Videopaměť',
        videoMemoryTip: 'Paměť vyhrazená pro snímání videa. Není sdílena se zbytkem systému.',
        videoMemoryGenerations_one: '{{count}} dřívější relace NanoKVM drží videopaměť',
        videoMemoryGenerations_few: '{{count}} dřívější relace NanoKVM drží videopaměť',
        videoMemoryGenerations_other: '{{count}} dřívějších relací NanoKVM drží videopaměť',
        videoMemoryReboot: 'Pro její uvolnění restartujte.',
        community: 'Komunita',
        hostname: 'Název hostitele',
        hostnameUpdated: 'Název hostitele byl aktualizován. Pro použití restartujte.',
        ipType: {
          Wired: 'Kabelové',
          Wireless: 'Bezdrátové',
          Other: 'Jiné'
        }
      },
      appearance: {
        title: 'Vzhled',
        display: 'Zobrazení',
        language: 'Jazyk',
        languageDesc: 'Vyberte jazyk rozhraní',
        webTitle: 'Název webu',
        webTitleDesc: 'Přizpůsobte název webové stránky',
        menuBar: {
          title: 'Panel nabídek',
          mode: 'Režim zobrazení',
          modeDesc: 'Zobrazení panelu nabídek na obrazovce',
          modeOff: 'Vypnuto',
          modeAuto: 'Automatické skrytí',
          modeAlways: 'Vždy viditelné',
          keyboardLedStatus: 'Indikátory zámku klávesnice',
          keyboardLedStatusDesc:
            'Zobrazit stav Num Lock, Caps Lock a Scroll Lock vzdáleného počítače',
          icons: 'Ikony podnabídky',
          iconsDesc: 'Zobrazení ikon podnabídky na liště nabídek'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Stav zámků vzdálené klávesnice',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Zapnuto',
        off: 'Vypnuto',
        unknown: 'Neznámé'
      },
      device: {
        title: 'Zařízení',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'Jas OLED',
          brightnessDescription: 'Nižší úroveň prodlužuje životnost displeje',
          brightnessLevels: {
            '64': 'Nejnižší',
            '96': 'Nízký',
            '128': 'Střední',
            '160': 'Vysoký',
            '207': 'Výchozí',
            '255': 'Maximální'
          },
          0: 'Nikdy',
          15: '15 sec',
          30: '30 sec',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 hodina'
        },
        ssh: {
          description: 'Povolit vzdálený přístup SSH',
          tip: 'Před povolením nastavte silné heslo (Účet – Změnit heslo)'
        },
        advanced: 'Pokročilá nastavení',
        cpuFreq: {
          title: 'Frekvence CPU',
          description: 'Nastavte takt CPU použitý při příštím spuštění',
          tip: 'CPU se spouští na 850 MHz a je dimenzován na 1000 MHz. Nová hodnota se použije při příštím spuštění, ne za běhu systému. 1000 MHz je v rámci specifikace; teplota je při obou nastaveních hluboko pod limity.',
          running: 'Aktuálně: {{mhz}} MHz',
          rebootToApply: 'pro použití restartujte',
          rebootConfirm: 'Restartovat nyní a použít {{mhz}} MHz?'
        },
        swap: {
          title: 'Vyměnit',
          disable: 'Zakázat',
          description: 'Nastavte velikost odkládacího souboru',
          tip: 'Povolení této funkce může zkrátit životnost vaší SD karty!'
        },
        zram: {
          title: 'Komprimovaný swap (zram)',
          description: 'Swap v komprimované RAM místo na SD kartě',
          tip: 'zram drží swap mimo SD kartu, takže ji neopotřebovává. Za ním není žádný diskový swap: pokud se zram zaplní, jádro ukončí některý proces, místo aby pomalu stránkovalo. Limit paměti určuje, kolik RAM může zram zabrat.',
          unavailable: 'Moduly jádra nejsou na tomto zařízení nainstalovány',
          inactive: 'Povoleno, ale zařízení se nespustilo',
          active: 'Aktivní - {{used}} z {{total}}, {{ratio}}x',
          off: 'Vypnuto',
          detail: {
            algorithm: 'Algoritmus: {{algorithm}}',
            memory: 'Využitá paměť: {{used}} z {{limit}}',
            memoryNoLimit: 'Využitá paměť: {{used}}, bez limitu',
            counters:
              'Stránky načtené ze swapu {{in}}, zapsané do swapu {{out}} (všechna swap zařízení, od spuštění)'
          }
        },
        mouseJiggler: {
          title: 'Mouse Jiggler',
          description: 'Zabraňte spánku vzdáleného hostitele',
          disable: 'Zakázat',
          absolute: 'Absolutní režim',
          relative: 'Relativní režim'
        },
        mdns: {
          description: 'Povolit službu zjišťování mDNS',
          tip: 'Vypnutí, pokud to není potřeba'
        },
        hdmi: {
          description: 'Povolit výstup HDMI/monitor',
          idleTimeoutTitle: 'Časový limit nečinnosti snímání',
          idleTimeoutDescription: 'Zastavit snímání HDMI po době bez aktivních diváků',
          minutes: 'min'
        },
        autostart: {
          title: 'Nastavení automatického spuštění skriptů',
          description: 'Správa skriptů, které se spouštějí automaticky při spuštění systému',
          new: 'Nové',
          deleteConfirm: 'Opravdu chcete tento soubor smazat?',
          yes: 'Ano',
          no: 'Ne',
          scriptName: 'Název skriptu automatického spuštění',
          scriptContent: 'Obsah skriptu automatického spuštění',
          settings: 'Nastavení'
        },
        hidOnly: 'HID-Pouze režim',
        hidOnlyDesc: 'Zastavit emulaci virtuálních zařízení a zachovat pouze základní ovládání HID',
        disk: 'Virtuální disk',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Virtuální síť',
        networkDesc: 'Připojit virtuální síťovou kartu na vzdáleném hostiteli',
        usbNetwork: {
          description:
            'Soukromé síťové spojení se vzdáleným hostitelem přes kabel USB. Hostitel dostane adresu bez brány a bez DNS, takže se přes NanoKVM nedostane do vaší LAN.',
          off: 'Vypnuto',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (pro hostitele bez NCM)',
          rndis: 'RNDIS (již se nenabízí)',
          rndisNote: 'Toto spojení používá RNDIS, které se již nenabízí. Zvolte NCM nebo ECM.',
          subnet: 'Podsíť',
          subnetDesc:
            'Soukromá síť IPv4, /24 až /30. NanoKVM použije první adresu, hostitel druhou.',
          addresses: 'NanoKVM: {{board}}, hostitel: {{host}}',
          invalidSubnet: 'Zadejte podsíť, například 172.31.255.0/30.',
          apply: 'Použít',
          confirm: 'Znovu připojit zařízení USB?',
          reenumerate:
            'Použití znovu sestaví připojení USB. Hostitel na několik sekund ztratí klávesnici, myš a virtuální disk.'
        },
        audio: 'Virtuální reproduktor',
        audioDesc:
          'Zpřístupní vzdálenému hostiteli zvukovou kartu USB, abyste slyšeli jeho zvuk. Hostitel ji musí zvolit jako výstupní zařízení. Přepnutí znovu sestaví připojení USB.',
        audioNote: 'Zvuk je dostupný v obou režimech H.264 (WebRTC a Direct), ne v MJPEG',
        console: 'Sériová konzole',
        consoleDesc:
          'Zpřístupní vzdálenému hostiteli sériový port USB pro přihlášení do tohoto NanoKVM, když je síť nedostupná',
        consoleTip:
          'Kdokoli, kdo ovládá vzdáleného hostitele, dostane přihlašovací výzvu tohoto NanoKVM. Před povolením nastavte silné heslo (Účet – Změnit heslo).',
        endpoints: {
          title: 'Koncové body USB',
          used: 'Využito {{used}} z {{total}}',
          cost: 'využívá {{cost}}',
          needs: 'potřebuje {{cost}}',
          full: 'Nedostatek koncových bodů USB. Nejprve vypněte něco jiného.',
          inactive:
            'Zapnuto, ale neběží: řadiči USB došly koncové body. Vypněte jiné zařízení a toto se ihned spustí.',
          explain:
            'Řadič USB má pevný počet vstupních koncových bodů a tento údaj je počítá. Pokud je povoleno více zařízení, než se vejde, klávesnice a myš zůstanou zachovány a ostatní se vypnou.',
          error: 'Zařízení nelze kontaktovat. Zkuste to znovu.',
          fitTogether: 'Společně se vejdou: {{sets}}'
        },
        reboot: 'Restartujte',
        rebootDesc: 'Opravdu chcete restartovat NanoKVM?',
        okBtn: 'Ano',
        cancelBtn: 'Ne'
      },
      network: {
        title: 'Síť',
        wifi: {
          title: 'Wi-Fi',
          description: 'Nastavit Wi-Fi',
          apMode: 'Režim AP je povolen, připojte se k Wi-Fi naskenováním QR kódu',
          connect: 'Připojit Wi-Fi',
          connectDesc1: 'Zadejte SSID sítě a heslo',
          connectDesc2: 'Zadejte heslo pro připojení k této síti',
          disconnect: 'Opravdu chcete síť odpojit?',
          failed: 'Připojení se nezdařilo, zkuste to znovu.',
          ssid: 'Název',
          password: 'Heslo',
          joinBtn: 'Připojit',
          confirmBtn: 'OK',
          cancelBtn: 'Zrušit'
        },
        tls: {
          description: 'Povolit protokol HTTPS',
          tip: 'Upozornění: Použití HTTPS může zvýšit latenci, zejména v režimu videa MJPEG.',
          restarting: 'Restartuje se server zařízení, potrvá to asi dvě minuty...',
          waiting: 'Čeká se, až zařízení znovu odpoví...',
          waitingHttp:
            'Přepíná se zpět na http. Pokud se tato stránka neotevře sama, načtěte ji znovu.'
        },
        ethernet: {
          title: 'IP adresa',
          description: 'Nastavte, jak NanoKVM získává adresu v drátové síti',
          dhcp: 'DHCP',
          manual: 'Ručně',
          networkDetails: 'Podrobnosti o síti',
          interface: 'Rozhraní',
          ipAddress: 'IP adresa',
          subnetMask: 'Maska podsítě',
          router: 'Router',
          save: 'Použít',
          invalidAddress: 'Zadejte platnou IP adresu',
          invalidMask: 'Zadejte platnou masku podsítě, například 255.255.255.0 nebo 24',
          invalidRouter: 'Zadejte platnou adresu routeru',
          addressRequired: 'IP adresa je povinná',
          maskRequired: 'Maska podsítě je povinná',
          applyTitle: 'Změnit adresu NanoKVM?',
          applyWarning:
            'Spojení s touto stránkou se ztratí. NanoKVM použije novou adresu a čeká {{seconds}} sekund, než se k němu na této adrese dostanete. Tím se změna zachová. Pokud se k němu nic nedostane, NanoKVM obnoví předchozí nastavení.',
          applyConfirm: 'Použít',
          applyCancel: 'Zrušit',
          applyFailed: 'Adresu se nepodařilo použít',
          trialTitle: 'Čeká se na potvrzení',
          trialDhcp: 'NanoKVM žádá o adresu přes DHCP.',
          trialStatic: 'NanoKVM je nyní na adrese {{address}}.',
          trialInstruction:
            'Otevřete NanoKVM na jeho nové adrese a přihlaste se, pokud o to požádá. Tím se změna zachová. Pokud se k NanoKVM nic do {{seconds}} sekund nedostane, obnoví předchozí nastavení.',
          trialOpen: 'Otevřít novou adresu',
          trialKeep: 'Zachovat toto nastavení',
          trialKept: 'Nová adresa je uložena',
          trialKeepFailed: 'Nastavení se nepodařilo zachovat',
          trialGone: 'Změna už byla vrácena zpět. Zkuste to znovu.',
          unsaved: 'Neuložené změny'
        },
        dns: {
          title: 'DNS',
          description: 'Nastavit DNS servery pro NanoKVM',
          mode: 'Režim',
          dhcp: 'DHCP',
          manual: 'Ručně',
          add: 'Přidat DNS',
          save: 'Uložit',
          invalid: 'Zadejte platnou IP adresu',
          noDhcp: 'Momentálně není k dispozici žádné DHCP DNS',
          saved: 'Nastavení DNS uloženo',
          saveFailed: 'Nastavení DNS se nepodařilo uložit',
          unsaved: 'Neuložené změny',
          maxServers: 'Je povoleno maximálně {{count}} DNS serverů',
          dnsServers: 'DNS servery',
          dhcpServersDescription: 'DNS servery jsou automaticky získávány z DHCP',
          manualServersDescription: 'DNS servery lze upravit ručně',
          networkDetails: 'Podrobnosti sítě',
          interface: 'Rozhraní',
          ipAddress: 'IP adresa',
          subnetMask: 'Maska podsítě',
          router: 'Router',
          none: 'Žádné'
        }
      },
      vpn: {
        loading: 'Načítání...',
        okBtn: 'Ano',
        cancelBtn: 'Ne',
        restart: 'Restartovat {{name}}?',
        stop: 'Zastavit {{name}}?',
        stopDesc:
          'Démon se nyní zastaví. Spouštění při startu je samostatný přepínač a zůstane beze změny.',
        update: 'Aktualizovat {{name}} na {{version}}?',
        updateDesc: 'Pokud démon běží, restartuje se. Přihlášení zůstane zachováno.',
        notInstall: '{{name}} není nainstalován.',
        install: 'Instalovat',
        installing: 'Instaluje se',
        installFailed: 'Instalace se nezdařila',
        retry: 'Zkusit znovu',
        notRunning: '{{name}} neběží. Pro pokračování ho spusťte.',
        run: 'Spustit',
        boot: 'Spouštět při startu',
        bootDesc: 'Spustit {{name}} při startu KVM.',
        enable: 'Povolit {{name}}',
        control: 'Řídicí server',
        connected: 'Připojeno',
        disconnected: 'Nepřipojeno',
        deviceName: 'Název zařízení',
        deviceIP: 'IP zařízení',
        account: 'Účet',
        version: 'Verze',
        uptime: 'Doba běhu',
        peers: 'Uzly',
        noPeers: 'Zatím žádné uzly.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Paměť',
        daemonRss: 'Démon',
        group: 'Skupina doplňků',
        high: 'zpomaleno nad {{size}}',
        max: 'nad {{size}} ukončeno jádrem',
        noGroup: 'Na této desce není paměťová skupina doplňků.',
        uninstall: 'Odinstalovat {{name}}',
        uninstallDesc: 'Opravdu chcete odinstalovat {{name}}? Přihlášení na desce zůstane.',
        blocked:
          '{{other}} běží nebo se spouští při startu. Současně může běžet jen jedna VPN: nejprve zastavte {{other}} a vypněte jeho spouštění při startu.',
        swap: {
          title: 'Odkládací paměť',
          tip: 'Pokud démonu dochází paměť, zkuste povolit odkládací paměť. Výchozí velikost odkládacího souboru je 256MB a lze ji upravit v "Nastavení > Zařízení".'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Obnovte stránku a zkuste to znovu. Nebo zkuste instalaci manuálně',
        download: 'Stáhnout',
        package: 'instalační balíček',
        unzip: 'a rozbalit ho',
        upTailscale: 'Nahrajte Tailscale do adresáře NanoKVM /usr/bin/',
        upTailscaled: 'Nahrajte Tailscaled do adresáře NanoKVM /usr/sbin/',
        refresh: 'Obnovit stránku',
        notLogin:
          'Zařízení nebylo dosud spárováno. Přihlaste se prosím a spárujte toto zařízení s vaším účtem.',
        urlPeriod: 'Tento odkaz je platný po dobu 10 minut',
        login: 'Přihlášení',
        loginSuccess: 'Přihlášení úspěšné',
        logout: 'Odhlásit se',
        logoutDesc: 'Opravdu se chcete odhlásit?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Toto zařízení se zatím nepřipojilo k síti NetBird. Připojte se pomocí instalačního klíče nebo se přihlaste přes SSO.',
        setupKey: 'Instalační klíč',
        setupKeyPlaceholder: 'Vložte instalační klíč z ovládacího panelu NetBird',
        join: 'Připojit',
        or: 'nebo',
        sso: 'Přihlásit přes SSO',
        urlPeriod: 'Tento odkaz je platný 10 minut',
        loginSuccess: 'Přihlášení úspěšné',
        logout: 'Odregistrovat',
        logoutDesc:
          'Odregistrování odebere tento uzel z vašeho účtu NetBird a smaže zde jeho konfiguraci. Opětovné připojení vyžaduje instalační klíč nebo přihlášení přes SSO a uzel může dostat novou IP. Pokračovat?'
      },
      update: {
        title: 'Zkontrolovat aktualizaci',
        queryFailed: 'Nepodařilo se získat verzi',
        updateFailed: 'Aktualizace se nezdařila. Zkuste to prosím znovu.',
        isLatest: 'Máte nejnovější verzi.',
        available: 'Je dostupná aktualizace. Opravdu chcete aktualizovat?',
        updating: 'Aktualizace zahájena. Prosím čekejte...',
        confirm: 'Potvrdit',
        cancel: 'Zrušit',
        preview: 'Náhled aktualizací',
        previewDesc: 'Získejte včasný přístup k novým funkcím a vylepšením',
        previewTip:
          'Uvědomte si prosím, že předběžné verze mohou obsahovat chyby nebo neúplné funkce!',
        customServer: {
          title: 'Vlastní aktualizační server',
          desc: 'Vyhledávejte a stahujte online aktualizace ze zadaného serveru',
          invalidUrl:
            'Zadejte platnou adresu adresáře serveru HTTP nebo HTTPS bez parametrů, fragmentu nebo souboru latest.json.',
          loadFailed: 'Konfiguraci aktualizačního serveru se nepodařilo načíst.',
          saveFailed: 'Konfiguraci aktualizačního serveru se nepodařilo uložit.',
          saved: 'Konfigurace aktualizačního serveru byla uložena.',
          save: 'Uložit',
          confirmTitle: 'Použít vlastní aktualizační server?',
          confirmDesc:
            'SHA-512 pouze ověřuje, že balíček odpovídá manifestu poskytnutému tímto serverem. Neprokazuje, že je balíček oficiálním vydáním NanoKVM. Vadný nebo škodlivý server může způsobit nefunkčnost zařízení, ztrátu dat nebo narušení zabezpečení systému.',
          confirm: 'Přesto použít',
          useSipeed: 'Použít oficiální server Sipeed',
          previewDisabled:
            'Testovací aktualizace nejsou při použití vlastního aktualizačního serveru dostupné.'
        },
        offline: {
          title: 'Offline aktualizace',
          desc: 'Aktualizace prostřednictvím místního instalačního balíčku',
          upload: 'Nahrát',
          checksumPlaceholder: 'Kontrolní součet SHA-256 (volitelný)',
          invalidChecksum: 'Kontrolní součet SHA-256 musí obsahovat 64 hexadecimálních znaků.',
          checksumMismatch: 'Ověření SHA-256 se nezdařilo. Balíček může být poškozený.',
          invalidName: 'Neplatný formát souboru. Stáhněte si prosím z vydání GitHubu.',
          updateFailed: 'Aktualizace se nezdařila. Zkuste to prosím znovu.'
        }
      },
      account: {
        title: 'Účet',
        webAccount: 'Název webového účtu',
        role: 'Role',
        roles: { admin: 'Správce', user: 'Uživatel' },
        password: 'Heslo',
        updateBtn: 'Update',
        logoutBtn: 'Odhlásit',
        logoutDesc: 'Opravdu se chcete odhlásit?',
        okBtn: 'Ano',
        cancelBtn: 'Ne',
        users: {
          title: 'Uživatelé',
          create: 'Vytvořit uživatele',
          enabled: 'Povolen',
          disabled: 'Zakázán',
          deviceOwner: 'Vlastník zařízení',
          resetPassword: 'Resetovat heslo',
          delete: 'Smazat',
          deleteConfirm: 'Smazat tohoto uživatele a zrušit všechny jeho relace?',
          created: 'Uživatel vytvořen',
          deleted: 'Uživatel smazán',
          passwordUpdated: 'Heslo aktualizováno',
          loadFailed: 'Uživatele se nepodařilo načíst',
          saveFailed: 'Uživatele se nepodařilo uložit',
          deleteFailed: 'Uživatele se nepodařilo smazat'
        }
      },
      apiKeys: {
        title: 'Klíče API',
        description:
          'Klíč jedná jménem svého vlastníka s rolí tohoto uživatele. Posílejte ho jako Authorization: Bearer <key> pro metriky a API, nebo jako X-Auth-Token pro Redfish.',
        name: 'Název',
        namePlaceholder: 'K čemu klíč slouží, např. prometheus',
        nameRequired: 'Pojmenujte klíč',
        nameTooLong: 'Název může mít nejvýše 64 znaků',
        unnamed: '(bez názvu)',
        create: 'Vytvořit klíč',
        created: 'Vytvořeno',
        owner: 'Vlastník',
        empty: 'Žádné klíče API',
        newKeyTitle: 'Váš nový klíč API',
        newKeyWarning:
          'Zkopírujte klíč nyní. Neukládá se a nelze ho znovu zobrazit. Pokud ho ztratíte, odvolejte ho a vytvořte nový.',
        copy: 'Kopírovat',
        copied: 'Zkopírováno',
        copyFailed: 'Kopírování se nezdařilo. Zkopírujte ručně.',
        done: 'Hotovo',
        revoke: 'Odvolat',
        revokeConfirmTitle: 'Odvolat tento klíč API?',
        revokeConfirmDesc: 'Vše, co používá "{{name}}", okamžitě přestane fungovat.',
        revoked: 'Klíč API odvolán',
        loadFailed: 'Klíče API se nepodařilo načíst',
        createFailed: 'Klíč API se nepodařilo vytvořit',
        revokeFailed: 'Klíč API se nepodařilo odvolat',
        cancelBtn: 'Zrušit'
      }
    },
    picoclaw: {
      title: 'PicoClaw asistent',
      empty: 'Otevřete panel a spusťte úlohu.',
      inputPlaceholder: 'Popište, co chcete, aby PicoClaw dělal',
      newConversation: 'Nová konverzace',
      processing: 'Zpracovává se...',
      agent: {
        defaultTitle: 'Obecný asistent',
        defaultDescription: 'Obecná nápověda pro chat, vyhledávání a pracovní prostor.',
        kvmTitle: 'Vzdálené ovládání',
        kvmDescription: 'Ovládejte vzdáleného hostitele prostřednictvím NanoKVM.',
        switched: 'Role agenta změněna',
        switchFailed: 'Přepnutí role agenta se nezdařilo'
      },
      send: 'Odeslat',
      cancel: 'Zrušit',
      status: {
        connecting: 'Připojování k bráně...',
        connected: 'Relace PicoClaw připojena',
        disconnected: 'Relace PicoClaw uzavřena',
        stopped: 'Požadavek na zastavení byl odeslán',
        runtimeStarted: 'Spuštěno běhové prostředí PicoClaw',
        runtimeStartFailed: 'Selhalo spuštění běhového prostředí PicoClaw',
        runtimeStopped: 'Běhové prostředí PicoClaw zastaveno',
        runtimeStopFailed: 'Zastavení běhového prostředí PicoClaw se nezdařilo',
        controlSwitchedToMCP: 'Ovládání bylo přepnuto na externí službu MCP'
      },
      connection: {
        runtime: {
          checking: 'Kontrola',
          restoring: 'Restoring PicoClaw',
          ready: 'Běhové prostředí připraveno',
          stopped: 'Běhové prostředí zastaveno',
          blockedByMCP: 'Externí ovládání MCP je aktivní',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Běhové prostředí není k dispozici',
          configError: 'Chyba konfigurace'
        },
        transport: {
          connecting: 'Připojování',
          connected: 'Připojeno',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'Nečinný',
          busy: 'Zaneprázdněn'
        }
      },
      message: {
        toolAction: 'Akce',
        observation: 'Pozorování',
        screenshot: 'Snímek obrazovky'
      },
      overlay: {
        locked: 'PicoClaw ovládá zařízení. Ruční zadávání je pozastaveno.'
      },
      control: {
        picoclaw: 'Ovládání zařízení: PicoClaw',
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Ovládání zařízení: externí MCP',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Ovládání zařízení: vypnuto',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Předat ovládání',
        release: 'Uvolnit',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'Ovládání PicoClaw povoleno',
        released: 'Ovládání PicoClaw uvolněno',
        grantFailed: 'Nepodařilo se předat ovládání PicoClaw',
        releaseFailed: 'Nepodařilo se uvolnit ovládání PicoClaw',
        grantConfirmTitle: 'Přepnout ovládání zařízení na PicoClaw?',
        grantConfirmDesc: 'Zápisy zařízení z externího MCP budou přerušeny.'
      },
      install: {
        install: 'Nainstalovat PicoClaw',
        installing: 'Instalace PicoClaw',
        success: 'PicoClaw úspěšně nainstalováno',
        failed: 'Nepodařilo se nainstalovat PicoClaw',
        uninstalling: 'Odinstalování běhového prostředí...',
        uninstalled: 'Běhové prostředí bylo úspěšně odinstalováno.',
        uninstallFailed: 'Odinstalace se nezdařila.',
        requiredTitle: 'PicoClaw není nainstalováno',
        requiredDescription: 'Nainstalujte PicoClaw před spuštěním běhového prostředí PicoClaw.',
        progressDescription: 'PicoClaw se stahuje a instaluje.',
        stages: {
          preparing: 'Příprava',
          downloading: 'Stahování',
          extracting: 'Extrakce',
          verifying: 'Ověřování',
          installing: 'Instalace probíhá',
          installed: 'Instalováno',
          install_timeout: 'Vypršel časový limit',
          install_failed: 'Selhalo'
        }
      },
      model: {
        requiredTitle: 'Je vyžadována konfigurace modelu',
        requiredDescription: 'Před použitím chatu PicoClaw nakonfigurujte model PicoClaw.',
        docsTitle: 'Průvodce konfigurací',
        docsDesc: 'Podporované modely a protokoly',
        menuLabel: 'Konfigurace modelu',
        modelIdentifier: 'Identifikátor modelu',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Klíč API',
        apiKeyPlaceholder: 'Zadejte klíč API modelu',
        save: 'Uložit',
        saving: 'Ukládání',
        saved: 'Konfigurace modelu uložena',
        saveFailed: 'Nepodařilo se uložit konfiguraci modelu',
        invalid: 'Identifikátor modelu, API Base URL a klíč API jsou povinné'
      },
      uninstall: {
        menuLabel: 'Odinstalovat',
        confirmTitle: 'Odinstalovat PicoClaw',
        confirmContent:
          'Opravdu chcete odinstalovat PicoClaw? Tím smažete spustitelný soubor a všechny konfigurační soubory.',
        confirmOk: 'Odinstalovat',
        confirmCancel: 'Zrušit'
      },
      history: {
        title: 'Historie',
        loading: 'Načítání relací...',
        emptyTitle: 'Zatím žádná historie',
        emptyDescription: 'Zde se zobrazí předchozí relace PicoClaw.',
        loadFailed: 'Nepodařilo se načíst historii relace',
        deleteFailed: 'Smazání relace se nezdařilo',
        deleteConfirmTitle: 'Smazat relaci',
        deleteConfirmContent: 'Opravdu chcete smazat "{{title}}"?',
        deleteConfirmOk: 'Smazat',
        deleteConfirmCancel: 'Zrušit',
        messageCount_one: '{{count}} zpráva',
        messageCount_few: '{{count}} zprávy',
        messageCount_other: '{{count}} zpráv',
        messageCount: '{{count}} zpráv'
      },
      config: {
        startRuntime: 'Spustit PicoClaw',
        stopRuntime: 'Zastavit PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Přepnout ovládání na PicoClaw?',
        enableConfirmDesc: 'Spuštěním PicoClaw se deaktivuje externí služba MCP.',
        enableConfirmOk: 'Spustit PicoClaw',
        enableConfirmCancel: 'Zrušit',
        title: 'Spustit PicoClaw',
        description: 'Spusťte běhové prostředí a začněte používat asistenta PicoClaw.',
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'Narazili jsme na problém',
      refresh: 'Obnovit',
      panel: 'Tato část stránky přestala fungovat',
      retry: 'Zkusit znovu'
    },
    fullscreen: {
      toggle: 'Přepnout na celou obrazovku'
    },
    input: {
      disconnected: 'Klávesnice a myš nejsou připojeny',
      disconnectedTls:
        'Prohlížeč bez dotazu odmítl zabezpečené spojení, které přenáší klávesnici a myš. Certifikát vygenerovaný tímto zařízením zatím není důvěryhodný. Otevřete tuto adresu na nové kartě, přijměte certifikát a pak stránku znovu načtěte. Spolehlivým řešením je certifikát nainstalovat.',
      disconnectedNever:
        'Spojení, které přenáší klávesnici a myš, se nepodařilo otevřít. Zbytek stránky funguje, protože ho nepoužívá. Zkontrolujte, zda ho nic mezi vámi a zařízením neblokuje.',
      disconnectedDropped:
        'Spojení, které přenáší klávesnici a myš, bylo přerušeno a neobnovilo se. Po restartu se připojí samo; pokud tento stav trvá, načtěte stránku znovu.',
      hidDisabled: 'HID je na tomto zařízení vypnuto (/boot/disable_hid).',
      keyFailed: 'Klávesu se nepodařilo odeslat.'
    },
    speaker: { title: 'Reproduktor', unmute: 'Zapnout zvuk', mute: 'Ztlumit' },
    menu: {
      collapse: 'Sbalit nabídku',
      expand: 'Rozbalte nabídku'
    },
    ion: {
      checking: 'Kontrola videopaměti před spuštěním streamu...',
      warn: 'Videopaměti je málo. Jeden restart serveru by ji vyčerpal. Až se vám to bude hodit, restartujte.',
      criticalTitle: 'Nedostatek videopaměti ke spuštění streamu',
      criticalBody:
        'Spuštění videa by vyčerpalo vyhrazenou paměť a zastavilo server. Všechny ostatní funkce fungují dál, včetně ovládání napájení a restartu. Tuto paměť uvolní jen restart NanoKVM.',
      criticalContinue: 'Přesto spustit video',
      criticalReboot: 'Restartovat NanoKVM',
      criticalRebooting: 'Restartování...'
    }
  }
};

export default cz;
