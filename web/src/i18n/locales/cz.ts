const cz = {
  translation: {
    feedback: {
      enabled: '{{name}} zapnuto',
      disabled: '{{name}} vypnuto',
      failed: 'Požadavek selhal. Zkuste to znovu.',
      network: 'Zařízení není dostupné. Zkontrolujte připojení a zkuste to znovu.',
      saved: 'Uloženo',
      timeout: 'Zařízení odpovídalo příliš dlouho. Zkuste to znovu.'
    },
    common: {
      copy: 'Kopírovat',
      copied: 'Zkopírováno',
      copyFailed: 'Kopírování se nezdařilo. Označte text a zkopírujte jej ručně.',
      notUpdating: 'Neaktualizuje se: poslední obnovení selhalo.',
      off: 'Vypnuto',
      running: 'Běží',
      save: 'Uložit',
      cancel: 'Zrušit',
      delete: 'Smazat',
      remove: 'Odebrat'
    },
    head: {
      desktop: 'Vzdálená plocha',
      login: 'Přihlášení',
      changePassword: 'Změna hesla',
      terminal: 'Terminál',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Heslo změněno. Přihlaste se novým heslem.',
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
          'To reset the passwords, pressing and holding the BOOT button on the IronKVM for 10 seconds.',
        reset3: 'Výchozí webový účet:',
        reset4: 'Výchozí účet SSH:',
        change1: 'Upozorňujeme, že tato akce změní následující hesla:',
        change2: 'Heslo pro webové přihlášení',
        change3: 'Heslo systémového uživatele root (heslo pro přihlášení SSH)',
        change4: 'Chcete-li hesla resetovat, podržte tlačítko BOOT na IronKVM.',
        resetDocs: 'Podrobný postup najdete v dokumentaci k hardwaru:',
        hardwareDocs: 'Wiki Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Nastavit Wi-Fi pro IronKVM',
      success: 'Please check the network status of IronKVM and visit the new IP address.',
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
      },
      ssidRequired: 'Zadejte název sítě, nejvýše 32 znaků',
      passwordLength: 'Heslo má 8 až 63 znaků. U otevřené sítě ho nechte prázdné.',
      passwordOptional: 'Heslo (prázdné pro otevřenou síť)',
      lost: 'Deska přestala odpovídat. Možná se připojila k síti a vypnula svůj konfigurační hotspot. Pokud se hotspot znovu objeví, připojení selhalo: připojte se k němu znovu a zkuste to znovu.',
      done: 'Nastavení dokončeno. Připojte toto zařízení zpět k obvyklé síti a otevřete desku na její nové adrese.'
    },
    screen: {
      viewOnly: 'Jen sledovat',
      viewOnlyTip:
        'Tato karta přestane posílat hostiteli klávesnici a myš. Skripty, jiggler myši a ostatní diváci nejsou ovlivněni.',
      viewOnlyOff: 'Vypnout jen sledování',
      viewOnlyBlocked: 'Je zapnuto jen sledování, hostiteli nebylo nic odesláno',
      pauseHidden: 'Pozastavit při skryté kartě',
      pauseHiddenTip:
        'Zastaví video i zvuk několik sekund po skrytí této karty a po návratu je znovu spustí.',
      screenshot: 'Snímek obrazovky',
      screenshotTip: 'Uloží obrazovku hostitele jako PNG v plné velikosti záznamu.',
      screenshotFailed: 'Snímek obrazovky se nezdařil',
      stream: {
        ok: 'obraz v pořádku',
        noSignal: 'bez signálu',
        failed: 'stream selhal'
      },
      codecNoWebrtcHevc: 'Tento prohlížeč neumí přijímat H.265 přes WebRTC',
      codecNoHevc: 'Tento prohlížeč neumí dekódovat H.265',
      codecNote:
        'Deska má jediný kodér, takže se tím změní stream pro všechny diváky. Pro použití v běžící relaci WebRTC se znovu připojte.',
      codec: 'Kodek',
      updateFailed: 'Nastavení nebylo použito',
      scale: 'Měřítko',
      title: 'Obrazovka',
      video: 'Režim videa',
      videoDirectTips: 'Chcete-li používat tento režim, povolte HTTPS v "Nastavení > Zařízení"',
      resolution: 'Rozlišení',
      ocr: {
        title: 'Přečíst text (OCR)',
        tips: 'Text se rozpoznává v tomto prohlížeči. Před zkopírováním jej můžete opravit.',
        hint: 'Táhněte přes text, který chcete přečíst. Stisknutím Esc akci zrušíte.',
        noPicture: 'Počkejte na video a pak táhněte přes text, který chcete přečíst.',
        cancel: 'Zrušit',
        language: 'Jazyk',
        languages: {
          eng: 'Angličtina'
        },
        preview: 'Vybraná oblast',
        capturing: 'Zachytávání obrazovky...',
        loading: 'Načítání rozpoznávání textu...',
        recognizing: 'Čtení textu...',
        noText: 'Ve vybrané oblasti nebyl nalezen žádný text.',
        copy: 'Kopírovat',
        copied: 'Zkopírováno do schránky',
        copyFailed: 'Kopírování do schránky se nezdařilo',
        selectAgain: 'Vybrat znovu',
        unsupported:
          'Tento prohlížeč nedokáže spustit rozpoznávání textu. Potřebuje WebAssembly SIMD, které mají současné prohlížeče.',
        captureFailed: 'Obrazovku se nepodařilo zachytit.',
        outside: 'Vybraná oblast je mimo obraz.',
        recognizeFailed: 'Rozpoznávání textu selhalo.'
      },
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
      qualityLossless: 'Nejlepší',
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
      },
      directConnectionFailed: 'Připojení videostreamu selhalo'
    },
    keyboard: {
      close: 'Zavřít',
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
        sendFailed: 'Neodesláno: vstupní spojení není k dispozici',
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
        saveFailed: 'Hlavní klávesu se nepodařilo uložit',
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
      jiggler: 'Pohyb myši proti uspání',
      keyJiggler: 'Stisk klávesy proti uspání',
      keyJigglerF15: 'Klávesa F15',
      keyJigglerShift: 'Klávesa Shift',
      keyJigglerCtrl: 'Klávesa Ctrl',
      keyJigglerF15Tip: 'F15 je nejméně rušivá: žádný běžný systém ani aplikace ji nepoužívá',
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
      touch: 'Dotykový režim',
      touchShort: 'Dotykový',
      absoluteStalled: 'Cílové zařízení ignoruje absolutní myš',
      absoluteStalledDesc:
        'Cílové zařízení přestalo přijímat hlášení absolutní myši, takže se pohyby kurzoru ztrácejí. Klávesnice není dotčena. Často pomůže obnovení USB; relativní režim používá jiný koncový bod.',
      useRelative: 'Přepnout na relativní režim',
      direction: 'Směr kolečka',
      scrollUp: 'Stejně jako tento počítač',
      scrollDown: 'Obrácené (přirozené posouvání)',
      speed: 'Rychlost kolečka',
      fast: 'Rychle',
      slow: 'Pomalu',
      requestPointer:
        'Používá se relativní režim. Klikněte prosím na plochu pro získání kurzoru myši.',
      resetHid: 'Resetovat HID',
      hidOnly: {
        switchFailed: 'Režim se nepodařilo přepnout. Zkontrolujte připojení a zkuste to znovu.',
        title: 'Režim pouze HID',
        desc: 'Pokud vaše myš a klávesnice přestanou reagovat a resetování HID nepomůže, může jít o problém s kompatibilitou mezi IronKVM a zařízením. Zkuste povolit režim HID-Only pro lepší kompatibilitu.',
        tip1: 'Povolení režimu HID-Only odpojí virtuální U-disk a virtuální síť',
        tip2: 'V režimu HID-Only je připojení obrazu zakázáno',
        rebuild: 'Přepnutí režimu znovu sestaví připojení USB. IronKVM se nerestartuje',
        enable: 'Povolit režim HID-Only',
        disable: 'Zakázat režim HID-Only'
      },
      resetHidDone: 'USB HID bylo resetováno',
      resetHidFailed: 'Reset USB HID se nezdařil'
    },
    image: {
      driveLoaded: 'obraz vložen',
      driveWarning: 'zkontrolujte varování',
      warning: {
        missing: 'Soubor obrazu byl smazán. Hostitel čte starou kopii, dokud ji nevysunete.',
        writable: 'Čtení i zápis: hostitel může tento obraz měnit.',
        tooBigForCd: 'Příliš velký pro CD mechaniku ({{size}}, limit {{max}}). Použijte disk.',
        tooSmallForCd: 'Příliš malý pro CD mechaniku ({{size}}). Použijte disk.',
        empty: 'Soubor je prázdný, nejspíš po nezdařeném nahrání nebo stažení.'
      },
      delete: 'Smazat',
      inUse: 'Používá se. Před smazáním jej vysuňte.',
      retry: 'Zkusit znovu',
      loadFailed: 'Seznam obrazů se nepodařilo načíst',
      readOnlyLocked: 'Chcete-li to změnit, vysuňte disk. Uplatní se při vložení obrazu.',
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
      deleteFailed: 'Smazání selhalo',
      ventoy: {
        statusNoKernel: 'Tento firmware nepodporuje',
        statusNotInstalled: 'Nenainstalováno',
        statusReady: 'Připraveno',
        statusSelected: 'Vybrané obrazy: {{count}}',
        statusInDrive: 'V diskové jednotce, {{size}}',
        noKernel:
          'Jádro tohoto firmwaru nepodporuje device-mapper, takže Ventoy nelze použít, dokud nenainstalujete obraz s touto podporou.',
        installDesc: 'Spusťte hostitele z několika obrazů na jednom disku, bez jejich kopírování.',
        install: 'Nainstalovat',
        installing: 'Stahování Ventoy, asi 20 MB. Může to trvat několik minut.',
        needsData: 'Ventoy potřebuje obraz IronKVM s připojeným oddílem /data.',
        uninstall: 'Odinstalovat',
        uninstallConfirm: 'Odstranit soubory Ventoy?',
        noImages: 'Žádné obrazy pro disk Ventoy.',
        onDisk: 'Na disku Ventoy',
        missing: 'Chybí: {{file}}',
        remove: 'Odebrat z disku Ventoy',
        setHint: 'Sadu obrazů lze měnit jen tehdy, když disk Ventoy není v žádné jednotce.',
        useAsDisk: 'Použít jako virtuální disk',
        failed: 'Požadavek Ventoy selhal',
        secureBoot:
          'Se zapnutým Secure Boot musí hostitel jednou zaregistrovat klíč Ventoy v MokManager. Soubor klíče ENROLL_THIS_KEY_IN_MOKMANAGER.cer je na oddílu VTOYEFI.',
        readOnly:
          'Hostitel vidí disk jen pro čtení, takže perzistence Ventoy a ventoy.json na jednotce nefungují.'
      },
      tips: {
        title: 'Jak nahrát',
        usb1: 'Připojte IronKVM k vašemu počítači přes USB.',
        usb2: 'Ujistěte se, že je virtuální disk připojen (Nastavení - Virtuální disk).',
        usb3: 'Otevřete virtuální disk na vašem počítači a zkopírujte soubor s obrazem do kořenového adresáře virtuálního disku.',
        scp1: 'Ujistěte se, že jsou IronKVM a váš počítač ve stejné místní síti.',
        scp2: 'Otevřete terminál na vašem počítači a použijte příkaz SCP pro nahrání souboru s obrazem do adresáře /data na zařízení IronKVM.',
        scp3: 'Příklad: scp cesta-k-vašemu-obrazu root@ip-nanokvm:/data',
        tfCard: 'SD Karta',
        tf1: 'Tato metoda je podporována na systémech Linux',
        tf2: 'Vyjměte SD kartu z IronKVM (u plné verze nejprve rozložte krabičku).',
        tf3: 'Vložte SD kartu do čtečky karet a připojte ji k vašemu počítači.',
        tf4: 'Zkopírujte soubor s obrazem do adresáře /data na SD kartě.',
        tf5: 'Vložte SD kartu zpět do IronKVM.'
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
      close: 'Zavřít',
      empty: 'Zatím žádné skripty. Nahrajte soubor .sh nebo .py a spusťte ho na desce.',
      loadFailed: 'Skripty se nepodařilo načíst',
      uploaded: 'Skript nahrán',
      uploadFailed: 'Skript se nepodařilo nahrát',
      started: 'Skript spuštěn na pozadí',
      deleteFailed: 'Skript se nepodařilo smazat',
      waitLimit: 'Čeká se na dokončení skriptu, nejvýše {{minutes}} minut.',
      timedOut:
        'Skript běžel déle než {{minutes}} minut a stránka přestala čekat. Na desce může stále běžet.'
    },
    terminal: {
      invalidBaud: 'Tato přenosová rychlost není podporována.',
      invalidPort: 'Zadejte cestu k zařízení v /dev, například /dev/ttyS1.',
      invalidSettings: 'Neplatné nastavení sériového portu. Toto je vlastní shell desky.',
      disconnected: 'Odpojeno. Stisknutím Enter se znovu připojíte.',
      title: 'Terminál',
      nanokvm: 'Terminál IronKVM',
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
      no: 'Ne',
      yes: 'Ano',
      deleteConfirm: 'Smazat tuto uloženou adresu?',
      delete: 'Smazat',
      wake: 'Probudit',
      rename: 'Přejmenovat',
      showMac: 'Zobrazit adresu MAC',
      showName: 'Zobrazit název',
      requestFailed: 'Zařízení nebylo dostupné, příkaz nebyl odeslán',
      deleteFailed: 'Smazání se nezdařilo',
      renameFailed: 'Přejmenování se nezdařilo',
      title: 'Wake-on-LAN',
      sending: 'Odesílání příkazu...',
      sent: 'Příkaz odeslán',
      input: 'Zadejte prosím MAC adresu',
      ok: 'OK'
    },
    download: {
      uploadFailed: 'Nahrávání selhalo',
      uploadSuccess: 'Nahrávání dokončeno',
      uploading: 'Nahrávání: {{file}}',
      downloadingPercent: 'Stahování ({{percent}}): {{file}}',
      downloading: 'Stahování: {{file}}',
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
      cancelFailed: 'Stažení se nepodařilo zrušit',
      bootMenu: 'Zaváděcí nabídka (netboot.xyz)',
      bootMenuPresent: '{{file}} už je v zařízení se správným kontrolním součtem',
      bootMenuDesc: 'Stáhnout ISO netboot.xyz s ověřeným kontrolním součtem pro virtuální CD'
    },
    alerts: {
      title: 'Vyžaduje pozornost',
      temperature: {
        warning: 'Deska má {{celsius}} °C. Zkontrolujte, zda k ní proudí vzduch.',
        critical: 'Deska má {{celsius}} °C, to je příliš. Zajistěte chlazení nebo ji vypněte.'
      },
      storage: {
        warning: 'Na {{path}} zbývá jen {{available}} z {{total}}. Velké obrazy se nemusí vejít.',
        critical:
          'Na {{path}} zbývá jen {{available}}. Nahrávání, stahování a instalace doplňků selžou. Smažte nepotřebné obrazy.'
      },
      vpn: '{{name}} se má spouštět při startu, ale neběží, takže vzdálený přístup přes něj nefunguje.',
      openVpn: 'Otevřít nastavení VPN',
      stream:
        'Videostream selhal. Zkuste jiný režim videa v nabídce Obrazovka nebo obnovte stránku.'
    },
    power: {
      resetDesc: 'Okamžitě restartuje hostitele. Neuložená práce se ztratí.',
      powerShortDesc: 'Zapne hostitele, nebo požádá jeho OS o vypnutí (ACPI).',
      powerLongDesc: 'Vynutí vypnutí hostitele bez řádného ukončení.',
      hddLed: 'LED disku',
      hddActive: 'Aktivní',
      hddIdle: 'Nečinný',
      title: 'Napájení',
      showConfirm: 'Potvrzení',
      showConfirmTip: 'Ptát se před krátkým stiskem napájení. Reset a dlouhý stisk se ptají vždy.',
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
      ledConnectedFailed: 'Nastavení LED napájení se nepodařilo uložit',
      powerLongConfirm:
        'Držet tlačítko napájení {{seconds}} s? Tím se vypne napájení bez vypnutí systému.',
      done: 'Tlačítko stisknuto',
      failed: 'Stisk tlačítka se nezdařil'
    },
    settings: {
      title: 'Nastavení',
      nav: {
        system: 'Systém',
        network: 'Síť',
        access: 'Přístup',
        integrations: 'Integrace',
        boot: 'Spouštění a média',
        browser: 'Tento prohlížeč',
        search: 'Najít nastavení',
        noMatch: 'Žádné nastavení neodpovídá',
        locked: 'Probíhá operace. Ostatní stránky ani zavření nejsou dostupné, dokud neskončí.',
        vpnProvider: 'Poskytovatel VPN'
      },
      mcp: {
        keyNote:
          'MCP používá vlastní API klíč, zobrazený níže. Klíče ze stránky API klíče zde nefungují.',
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
        cancelBtn: 'Zrušit',
        showKey: 'Zobrazit klíč',
        hideKey: 'Skrýt klíč',
        regenerateKey: 'Vygenerovat nový klíč'
      },
      redfish: {
        example: 'Příklad',
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
      ipmi: {
        copyBeforeSave: 'Zkopírujte heslo hned. Po uložení už jej nelze zobrazit.',
        noLogin:
          'IPMI je zapnuté, ale žádný aktivní účet nemá heslo IPMI, takže se nikdo nepřihlásí. Nastavte ho níže.',
        title: 'IPMI',
        warning:
          'Ověřování IPMI je ze své podstaty slabé. Kdokoli, kdo se k desce dostane a zná uživatelské jméno, může získat hash IPMI hesla tohoto uživatele a pokusit se ho prolomit offline. Používejte generovaná hesla, zapínejte IPMI jen v důvěryhodné síti a kde to nástroj umí, dejte přednost Redfish přes HTTPS.',
        service: 'IPMI přes LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) na UDP portu 623 pro napájení a stav hostitele. IPMI 1.5 a cipher suite 0 jsou odmítnuty. Vypnutím se ukončí všechny relace IPMI.',
        example: 'Příklad',
        copyFailed: 'Kopírování selhalo. Zkopírujte ručně.',
        ledOn: 'K dispozici jsou stav napájení, on, off, soft, cycle a reset.',
        ledOff:
          '"LED napájení připojena" je v nabídce napájení vypnutá, takže stav napájení není znám. Funguje jen "power reset": status, on, off, soft a cycle jsou odmítnuty.',
        accounts: 'Účty',
        accountsDesc:
          'IPMI se přihlašuje účty KVM, každý s vlastním IPMI heslem, odlišným od webového hesla. Správci dostanou ADMINISTRATOR. Uživatelé dostanou USER: s "-L USER" mohou číst stav napájení, ale ne ho měnit.',
        passwordSet: 'IPMI heslo nastaveno',
        passwordNotSet: 'Bez IPMI hesla: nelze se přihlásit přes IPMI',
        nameTooLong: 'Jméno je delší než 16 znaků, což IPMI nedovoluje',
        accountDisabled: 'Účet je zakázán',
        setPassword: 'Nastavit heslo',
        changePassword: 'Změnit heslo',
        remove: 'Odebrat',
        removeConfirmTitle: 'Odebrat IPMI heslo účtu {{user}}?',
        removeConfirmDesc: 'Účet se už nebude moci přihlásit přes IPMI a jeho relace IPMI skončí.',
        passwordTitle: 'IPMI heslo pro {{user}}',
        passwordDesc:
          '12 až 20 tisknutelných znaků ASCII, odlišné od webového hesla. IPMI vyžaduje, aby deska uchovávala heslo v podobě, kterou dokáže znovu přečíst, proto použijte heslo, které nepoužíváte nikde jinde. Před uložením si ho zkopírujte: znovu se nezobrazí.',
        passwordPlaceholder: 'IPMI heslo',
        generate: 'Vygenerovat',
        copy: 'Kopírovat',
        save: 'Uložit',
        passwordLength: 'Použijte 12 až 20 znaků.',
        passwordChars: 'Použijte jen tisknutelné znaky ASCII.',
        saved: 'IPMI heslo uloženo',
        failed: 'Operace IPMI selhala',
        okBtn: 'Potvrdit',
        cancelBtn: 'Zrušit'
      },
      ssh: {
        service: 'SSH server',
        serviceDesc: 'Spustit sshd nyní a při každém startu',
        failed: 'Nastavení SSH se nepodařilo načíst',
        rootDefault: 'root má stále tovární heslo',
        rootEmpty: 'root nemá heslo',
        rootWarning:
          'Kdokoli se dostane ke konzoli nebo SSH, se může přihlásit jako root. Nastavte heslo v {{account}} > {{password}}: pro vlastníka zařízení nastaví i heslo uživatele root.',
        connection: 'Připojení',
        command: 'Přihlásit se jako root',
        port: 'Port',
        viaVpn: 'Přes {{name}}',
        notRunning: 'sshd neběží. Pro připojení zapněte SSH server.',
        hostKeys: 'Otisky klíčů hostitele',
        hostKeysDesc: 'Porovnejte je s tím, co ssh ukáže při prvním připojení.',
        noHostKeys: 'Zatím žádné klíče hostitele. sshd je vytvoří při prvním spuštění.',
        keys: 'Autorizované klíče',
        keysDesc:
          'Veřejné klíče, které se mohou přihlásit jako root. Jsou uložené na datovém oddílu, takže přežijí aktualizace.',
        noKeys: 'Zatím žádné autorizované klíče.',
        noComment: 'bez komentáře',
        addPlaceholder: 'Vložte jeden veřejný klíč, například obsah ~/.ssh/id_ed25519.pub',
        add: 'Přidat klíč',
        added: 'Klíč přidán',
        removed: 'Klíč odebrán',
        deleteConfirm: 'Odebrat tento klíč?',
        deleteConfirmDesc: 'Už se nebude moci přihlásit. Otevřené relace zůstanou otevřené.',
        invalidKey: 'Toto není veřejný klíč. Vložte jeden řádek ze souboru .pub.',
        keyOptions: 'Klíče s volbami jako command= nebo from= zde nejsou přijímány.',
        duplicateKey: 'Tento klíč už je autorizován.',
        lastKey: 'Poslední klíč nelze odebrat, dokud je zapnuté přihlašování pouze klíči.',
        keysOnly: 'Pouze klíče',
        keysOnlyDesc:
          'Vypnout přihlašování heslem a keyboard-interactive. Otevřené relace zůstanou otevřené.',
        keysOnlyNeedsKey: 'Nejprve přidejte autorizovaný klíč, jinak by se nikdo nepřihlásil.',
        keysOnlyOn: 'Přihlašování heslem vypnuto',
        keysOnlyOff: 'Přihlašování heslem zapnuto',
        notHonoured:
          'sshd v tomto obrazu toto nastavení nečte, přihlašování heslem tedy zůstává zapnuté.',
        reloadFailed:
          'Uloženo, ale sshd se nepodařilo znovu načíst. Projeví se při příštím spuštění sshd.',
        notApplied:
          'sshd stále přijímá hesla. Vypněte a zapněte SSH server, aby se nastavení projevilo.',
        changePort: 'Změnit',
        portConfirm: 'Změnit port SSH na {{port}}?',
        portConfirmDesc:
          'Aktuální relace SSH zůstanou otevřené. Nová připojení musí používat port {{port}}. Ověřte, že to firewall povoluje.',
        portChanged: 'Port SSH změněn na {{port}}',
        portInvalid: 'Zadejte port od 1 do 65535.',
        portReserved: 'Tento port používá samotné IronKVM. Zvolte jiný.',
        portInUse: 'Na tomto portu už na IronKVM naslouchá jiný program.',
        portNotHonoured: 'sshd v tomto obrazu toto nastavení nečte, port zůstává beze změny.'
      },
      vnc: {
        address: 'Adresa',
        certHint:
          'VeNCrypt X509Plain používá certifikát zařízení podepsaný sám sebou, takže klient při prvním připojení varuje. Přijměte ho, nebo certifikát uložte z HTTPS adresy této stránky a předejte ho TigerVNC pomocí -X509CA=<soubor>.',
        title: 'VNC',
        service: 'Server VNC',
        serviceDesc:
          'Umožní klientovi VNC, například TigerVNC nebo Remmina, zobrazit a ovládat hostitele. Klient musí podporovat kódování Tight. Vždy jedna relace.',
        credentials:
          'Přihlaste se účtem KVM. Spojení je šifrováno certifikátem TLS desky (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'Port TCP, na kterém server naslouchá.',
        maxFps: 'Limit snímků',
        maxFpsDesc: 'Nejvyšší počet snímků za sekundu, který klient dostane.',
        vncAuth: 'Jednoduché ověřování VNC',
        vncAuthDesc: 'Pro klienty bez VeNCrypt. Kontroluje samostatné heslo VNC místo účtu.',
        vncAuthWarning:
          'Jednoduché ověřování VNC spojení nešifruje. Kdokoli na síťové cestě uvidí obrazovku i stisky kláves. Používejte ho jen v důvěryhodné síti.',
        password: 'Heslo VNC',
        passwordSet: 'Heslo je nastaveno. Pro změnu zadejte nové.',
        passwordInvalid: 'Heslo VNC musí mít 6 až 8 znaků.',
        save: 'Uložit',
        saved: 'Nastavení uloženo',
        state: 'Stav',
        listening: 'Naslouchá na portu {{port}}',
        notListening: 'Nenaslouchá',
        noSession: 'Žádná otevřená relace',
        client: 'Klient',
        user: 'Uživatel',
        method: 'Ověřování',
        methodVencrypt: 'Účet přes TLS',
        methodVnc: 'Heslo VNC',
        since: 'Připojeno od',
        resolution: 'Rozlišení',
        framesSent: 'Odeslané snímky',
        lastError: 'Poslední relace skončila: {{error}}',
        refresh: 'Obnovit',
        disconnect: 'Odpojit',
        disconnectConfirmTitle: 'Ukončit relaci VNC?',
        disconnectConfirmDesc:
          'Klient se okamžitě odpojí a všechny držené klávesy a tlačítka se uvolní.',
        failed: 'Operace VNC selhala',
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
        actionReset: 'Resetovat',
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
        failed: 'Operace watchdogu selhala',
        powerNeedsLed: 'Cyklus napájení vyžaduje „LED napájení připojena“ v nabídce napájení.',
        noLedConfirmTitle: 'Zapnout watchdog bez LED napájení?',
        noLedConfirmDesc:
          'Deska nevidí, kdy je hostitel vypnutý, takže ho považuje za stále zapnutý. Pokud hostitele vypnete, watchdog po uplynutí časového limitu stiskne reset. Abyste tomu předešli, připojte LED napájení.',
        noLedConfirmOk: 'Zapnout',
        cancel: 'Zrušit'
      },
      media: {
        title: 'Virtuální média',
        description:
          'Nastavení dialogu Média na panelu nástrojů. Připojení obrazů, jejich přidání a výběr sady pro Ventoy probíhá v dialogu.',
        ejectFirst: 'Disk Ventoy je v jednotce. Pro odinstalaci jej vysuňte v dialogu Média.'
      },
      netboot: {
        title: 'Síťové spuštění',
        isoDownload: 'Stáhnout',
        description:
          'Spustit hostitele ze sítě: iPXE a nabídka obrazů na KVM přes síťové propojení USB, nebo netboot.xyz přes proxy DHCP v síti LAN.',
        addon: 'dnsmasq a zaváděcí soubory',
        addonDesc:
          'Instalováno na /data: dnsmasq z Alpine, iPXE a netboot.xyz z jejich vydání, každý ověřený kontrolním součtem.',
        install: 'Instalovat',
        installing: 'Probíhá instalace. Může to trvat několik minut.',
        uninstall: 'Odinstalovat',
        uninstallConfirm: 'Vypnout síťové spuštění a odstranit dnsmasq a zaváděcí soubory?',
        needsData: 'Síťové spuštění vyžaduje obraz IronKVM s připojeným oddílem /data.',
        usb: 'Na síťovém propojení USB',
        usbDesc:
          'Dokud je síťové propojení USB zapnuté, obsluhuje ho dnsmasq místo udhcpd. Hostitel dostane svou jedinou adresu bez směrovače a bez serveru DNS, iPXE pro svou architekturu a nabídku obrazů ISO na KVM.',
        linkOff: 'Síťové propojení USB je vypnuté. Zapněte ho v Zařízení, Síť USB.',
        menuUrl: 'Nabídka',
        leases: 'Zápůjčka hostitele',
        noLeases: 'Zatím žádná',
        netbootxyzNote:
          'netboot.xyz v nabídce se načítá z internetu, kam propojení USB nedosáhne. Hostitel potřebuje internet na jiném síťovém portu.',
        lan: 'Proxy DHCP v síti LAN',
        lanDesc:
          'Odpovídá klientům PXE v síti LAN nabídkou netboot.xyz, který si pak načte svou nabídku z internetu. Nikdy nepřiděluje adresy a nezpřístupňuje obrazy na KVM.',
        lanWarning:
          'netboot.xyz je nabídnut každému klientovi PXE v této síti LAN, nejen hostiteli. Zapínejte to jen v síti, kterou spravujete.',
        lanConfirm: 'Zapnout proxy DHCP v síti LAN?',
        lanInterface: 'LAN',
        running: 'Běží',
        stopped: 'Neběží',
        images: 'Obrazy v nabídce',
        noImages: 'V adresáři obrazů nejsou žádné obrazy ISO.',
        boots: 'Nedávná spuštění',
        noBoots: 'Hostitel zatím nic nestáhl.',
        log: 'Protokol dnsmasq',
        refresh: 'Obnovit',
        okBtn: 'Potvrdit',
        cancelBtn: 'Zrušit',
        failed: 'Operace síťového spuštění selhala'
      },
      about: {
        title: 'O IronKVM',
        information: 'Informace',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Verze aplikace',
        applicationTip: 'Verze webové aplikace IronKVM',
        image: 'Verze obrazu',
        imageTip: 'Obraz karty IronKVM a systémový obraz NanoKVM, na kterém je postaven',
        kernel: 'Verze jádra',
        kernelTip: 'Verze aktuálně běžícího jádra Linuxu',
        deviceKey: 'Klíč zařízení',
        videoMemory: 'Videopaměť',
        videoMemoryTip: 'Paměť vyhrazená pro snímání videa. Není sdílena se zbytkem systému.',
        videoMemoryGenerations_one: '{{count}} dřívější relace IronKVM drží videopaměť',
        videoMemoryGenerations_few: '{{count}} dřívější relace IronKVM drží videopaměť',
        videoMemoryGenerations_other: '{{count}} dřívějších relací IronKVM drží videopaměť',
        videoMemoryReboot: 'Pro její uvolnění restartujte.',
        community: 'Komunita',
        hostname: 'Název hostitele',
        hostnameUpdated: 'Název hostitele byl aktualizován. Pro použití restartujte.',
        ipType: {
          Wired: 'Kabelové',
          Wireless: 'Bezdrátové',
          Other: 'Jiné'
        },
        hostnameInvalid:
          'Použijte písmena, číslice a pomlčky, nejvýše 63 v každé části oddělené tečkou. Pomlčka nesmí být na začátku ani na konci části.',
        hostnameFailed: 'Název hostitele se nepodařilo změnit',
        editHostname: 'Upravit název hostitele',
        docs: 'Dokumentace',
        hardware: 'Hardware',
        hardwareFaq: 'Časté dotazy k hardwaru',
        disclaimer:
          'IronKVM: zabezpečený komunitní firmware pro Sipeed NanoKVM. Bez vazby na společnost Sipeed.',
        basedOn: 'založeno na NanoKVM {{version}}'
      },
      preferences: {
        title: 'Předvolby'
      },
      performance: {
        title: 'Výkon'
      },
      appearance: {
        thisBrowser: 'Tento prohlížeč',
        thisBrowserDesc: 'Uloženo pouze v tomto prohlížeči. Ostatní prohlížeče mají vlastní.',
        deviceWide: 'Zařízení',
        deviceWideDesc: 'Uloženo v zařízení. Platí pro každého, kdo ho otevře.',
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
          15: '15 s',
          30: '30 s',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 hodina'
        },
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Přední panel'
        },
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
          tip: 'Povolení této funkce může zkrátit životnost vaší SD karty!',
          active: 'Aktivní - {{used}} z {{total}}',
          inactive: 'Nastaveno, ale nepoužívá se'
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
        hidOnly: 'HID-Pouze režim',
        hidOnlyDesc: 'Zastavit emulaci virtuálních zařízení a zachovat pouze základní ovládání HID',
        disk: 'Virtuální disk',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Virtuální síť',
        networkDesc: 'Připojit virtuální síťovou kartu na vzdáleném hostiteli',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Hostitel:',
          description:
            'Soukromé síťové spojení se vzdáleným hostitelem přes kabel USB. Hostitel dostane adresu bez brány a bez DNS, takže se přes IronKVM nedostane do vaší LAN.',
          mode: 'Protokol',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (pro hostitele bez NCM)',
          rndis: 'RNDIS (již se nenabízí)',
          rndisNote: 'Toto spojení používá RNDIS, které se již nenabízí. Zvolte NCM nebo ECM.',
          subnet: 'Podsíť',
          subnetDesc:
            'Soukromá síť IPv4, /24 až /30. IronKVM použije první adresu, hostitel druhou.',
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
          'Zpřístupní vzdálenému hostiteli sériový port USB pro přihlášení do tohoto IronKVM, když je síť nedostupná',
        consoleTip:
          'Kdokoli, kdo ovládá vzdáleného hostitele, dostane přihlašovací výzvu tohoto IronKVM. Před povolením nastavte silné heslo (Účet – Změnit heslo).',
        usbApply: {
          changed: 'Změněno',
          discard: 'Zahodit',
          pending: 'Změny zatím nejsou použity.'
        },
        endpoints: {
          title: 'USB sloty',
          free: '{{free}} z {{total}} volných',
          slots: 'Sloty: {{count}}',
          full: 'Nedostatek volných USB slotů. Nejprve vypněte něco jiného.',
          inactive: 'Zapnuto, ale neběží: USB řadiči došly sloty. Vypněte jiné zařízení a toto se hned spustí.',
          explain: 'USB řadič má pevný počet slotů (vstupních endpointů) a klávesnice s myší vždy některé zabírají. Pokud je zapnuto více zařízení, než se vejde, klávesnice a myš zůstanou a ostatní se vypnou.',
          error: 'Zařízení nelze kontaktovat. Zkuste to znovu.',
          fitTogether: 'Společně se vejdou: {{sets}}'
        },
        reboot: 'Restartujte',
        rebootDesc: 'Opravdu chcete restartovat IronKVM?',
        okBtn: 'Ano',
        cancelBtn: 'Ne',
        rebootFailed: 'Restart se nezdařil'
      },
      network: {
        title: 'Síť',
        wifi: {
          disconnectBtn: 'Odpojit',
          disconnectWarning:
            'Pokud k IronKVM přistupujete přes tuto Wi-Fi síť, tato stránka ztratí spojení.',
          disconnected: 'Wi-Fi odpojena',
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
            'Přepíná se zpět na http. Pokud se tato stránka neotevře sama, načtěte ji znovu.',
          failed: 'Nastavení HTTPS se nepodařilo změnit',
          enableConfirm: 'Zapnout HTTPS?',
          disableConfirm: 'Vypnout HTTPS?',
          confirmDesc:
            'Tím se odhlásíte a server zařízení se restartuje, což trvá asi dvě minuty. Stránka pak otevře {{url}}.',
          confirmOk: 'Pokračovat',
          confirmCancel: 'Zrušit'
        },
        ethernet: {
          title: 'IP adresa',
          description: 'Nastavte, jak IronKVM získává adresu v drátové síti',
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
          applyTitle: 'Změnit adresu IronKVM?',
          applyWarning:
            'Spojení s touto stránkou se ztratí. IronKVM použije novou adresu a čeká {{seconds}} sekund, než se k němu na této adrese dostanete. Tím se změna zachová. Pokud se k němu nic nedostane, IronKVM obnoví předchozí nastavení.',
          applyConfirm: 'Použít',
          applyCancel: 'Zrušit',
          applyFailed: 'Adresu se nepodařilo použít',
          trialTitle: 'Čeká se na potvrzení',
          trialDhcp: 'IronKVM žádá o adresu přes DHCP.',
          trialStatic: 'IronKVM je nyní na adrese {{address}}.',
          trialInstruction:
            'Otevřete IronKVM na jeho nové adrese a přihlaste se, pokud o to požádá. Tím se změna zachová. Pokud se k IronKVM nic do {{seconds}} sekund nedostane, obnoví předchozí nastavení.',
          trialOpen: 'Otevřít novou adresu',
          trialKeep: 'Zachovat toto nastavení',
          trialKept: 'Nová adresa je uložena',
          trialKeepFailed: 'Nastavení se nepodařilo zachovat',
          trialGone: 'Změna už byla vrácena zpět. Zkuste to znovu.',
          unsaved: 'Neuložené změny'
        },
        dns: {
          title: 'DNS',
          description: 'Nastavit DNS servery pro IronKVM',
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
        connect: 'Připojit',
        connectDesc: 'Připojit se k síti {{name}}. Vypnuto odpojí bez zastavení služby.',
        kvmUrl: 'Adresa KVM',
        moreTip: 'Další akce',
        restartTip: 'Restartovat',
        stopTip: 'Zastavit',
        updateTip: 'Aktualizovat na {{version}}',
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
          tip: 'Pokud démonu dochází paměť, zkuste zapnout odkládací paměť. Nastavuje se v "Nastavení > Výkon".'
        },
        copy: 'Kopírovat',
        copied: 'Odkaz zkopírován',
        copyFailed: 'Odkaz se nepodařilo zkopírovat. Označte ho a zkopírujte ručně.',
        open: 'Otevřít',
        checkAgain: 'Zkontrolovat znovu',
        notSignedIn: 'Zatím nepřihlášeno. Dokončete přihlášení přes odkaz a zkontrolujte znovu.',
        checkFailed: 'Stav přihlášení se nepodařilo zjistit',
        loginWaiting: 'Stránka to kontroluje každých pár sekund a po přihlášení pokračuje.',
        uninstallFailed: 'Odinstalace se nezdařila',
        loginFailed: 'Přihlášení se nezdařilo'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Stáhnout',
        package: 'instalační balíček',
        unzip: 'a rozbalit ho',
        notLogin:
          'Zařízení nebylo dosud spárováno. Přihlaste se prosím a spárujte toto zařízení s vaším účtem.',
        urlPeriod: 'Tento odkaz je platný po dobu 10 minut',
        login: 'Přihlášení',
        logout: 'Odhlásit se',
        logoutDesc: 'Opravdu se chcete odhlásit?',
        manualIntro: 'Nebo jej nainstalujte ručně přes SSH:',
        copyBinaries: 'Zkopírujte tailscale a tailscaled do {{dir}} na IronKVM',
        linksFile: 'Ve stejném adresáři vytvořte soubor s názvem links s těmito dvěma řádky:',
        rebootRefresh: 'Restartujte IronKVM a poté obnovte tuto stránku'
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
        logout: 'Odregistrovat',
        logoutDesc:
          'Odregistrování odebere tento uzel z vašeho účtu NetBird a smaže zde jeho konfiguraci. Opětovné připojení vyžaduje instalační klíč nebo přihlášení přes SSO a uzel může dostat novou IP. Pokračovat?',
        joinFailed: 'Nepodařilo se připojit k síti'
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
            'SHA-512 pouze ověřuje, že balíček odpovídá manifestu poskytnutému tímto serverem. Neprokazuje, že je balíček oficiálním vydáním IronKVM. Vadný nebo škodlivý server může způsobit nefunkčnost zařízení, ztrátu dat nebo narušení zabezpečení systému.',
          confirm: 'Přesto použít',
          useSipeed: 'Použít oficiální server Sipeed',
          previewDisabled:
            'Testovací aktualizace nejsou při použití vlastního aktualizačního serveru dostupné.'
        },
        offline: {
          chooseFile: 'Vybrat soubor',
          installing: 'Nahráno. Instaluje se...',
          noFile: 'Není vybrán žádný soubor',
          title: 'Offline aktualizace',
          desc: 'Aktualizace prostřednictvím místního instalačního balíčku',
          upload: 'Nahrát',
          checksumPlaceholder: 'Kontrolní součet SHA-256 (volitelný)',
          invalidChecksum: 'Kontrolní součet SHA-256 musí obsahovat 64 hexadecimálních znaků.',
          checksumMismatch: 'Ověření SHA-256 se nezdařilo. Balíček může být poškozený.',
          invalidName: 'Neplatný formát souboru. Stáhněte si prosím z vydání GitHubu.',
          updateFailed: 'Aktualizace se nezdařila. Zkuste to prosím znovu.'
        },
        updateTo: 'Aktualizovat na {{version}}',
        updateConfirmDesc:
          'Zařízení nainstaluje aktualizaci a restartuje svůj server. Stránka se znovu načte, až bude server zpět.',
        releaseNotes: 'Poznámky k vydání'
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
        mcpNote: 'Tyto klíče nefungují pro MCP, který má vlastní klíč na stránce MCP.',
        metricsUrl: 'URL metrik',
        monitoring: 'Monitorování',
        monitoringDesc:
          'Prometheus čte metriky pomocí API klíče z této stránky, odeslaného jako Bearer token. Číst je může jakákoli role.',
        scrapeConfig: 'Konfigurace scrape pro Prometheus',
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
        kvmDescription: 'Ovládejte vzdáleného hostitele prostřednictvím IronKVM.',
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
          restoring: 'Obnovování PicoClaw',
          ready: 'Běhové prostředí připraveno',
          stopped: 'Běhové prostředí zastaveno',
          blockedByMCP: 'Externí ovládání MCP je aktivní',
          readyBlockedByMCP: 'Běhové prostředí běží, ale vstup zařízení nyní ovládá externí MCP.',
          readyWithoutControl:
            'Běhové prostředí běží. Před opětovným připojením předejte PicoClaw ovládání zařízení.',
          unavailable: 'Běhové prostředí není k dispozici',
          configError: 'Chyba konfigurace'
        },
        transport: {
          connecting: 'Připojování',
          connected: 'Připojeno',
          disconnected: 'Odpojeno',
          reconnect: 'Znovu připojit',
          reconnectDescription: 'Znovu se připojit k běžící relaci PicoClaw.',
          reconnectBlocked: 'PicoClaw potřebuje před opětovným připojením ovládání zařízení.'
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
        picoclawDescription:
          'PicoClaw může zadávat vstup z klávesnice a myši. Ruční vstup se může pozastavit.',
        mcp: 'Ovládání zařízení: externí MCP',
        mcpDescription: 'Externí MCP může zapisovat do zařízení. PicoClaw vstup nepřevezme.',
        off: 'Ovládání zařízení: vypnuto',
        offDescription:
          'AI nebude zadávat vstup z klávesnice ani myši. Ruční ovládání zůstává k dispozici.',
        transitioning: 'Ovládání zařízení: přepínání',
        transitioningDescription: 'Ovládání zařízení se synchronizuje. Počkejte prosím.',
        grant: 'Předat ovládání',
        release: 'Uvolnit',
        releasing: 'Uvolňování...',
        switching: 'Přepínání...',
        releasingLabel: 'Ovládání zařízení: uvolňování',
        releasingDescription: 'Ovládání zařízení se vrací. PicoClaw zastavil probíhající zápisy.',
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
        switchFromMCP: 'Přepnout na PicoClaw a spustit',
        takeoverAndStart: 'Převzít a spustit'
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
    speaker: {
      title: 'Reproduktor',
      unmute: 'Zapnout zvuk',
      mute: 'Ztlumit',
      hostIdle: 'Hostitel neposílá zvuk',
      hostIdleHint: 'Přehrajte něco na hostiteli nebo v něm zvolte KVM jako zvukový výstup.'
    },
    upstream: {
      check: 'Zkontrolovat aktualizace',
      updateTo: 'Aktualizovat na {{version}}',
      confirm: 'Aktualizovat {{name}} na {{version}}?',
      confirmDesc:
        'Nové vydání se stáhne z GitHubu a ověří podle kontrolních součtů, které zveřejňuje. Pokud cokoli selže, zůstane současná verze.',
      ok: 'Aktualizovat',
      upToDate: 'Aktuální',
      builtIn: 'vestavěná',
      checkFailed: 'Aktualizace nelze zkontrolovat: {{error}}',
      unverifiable: 'Verze {{version}} není nabízena: {{reason}}',
      inUse: 'Nyní nelze aktualizovat: {{reason}}',
      running: 'Aktualizace na {{version}}...',
      done: '{{name}} aktualizováno na {{version}}',
      failed: 'Poslední aktualizace selhala: {{error}}'
    },
    menu: {
      mediaAdd: 'Přidat obraz',
      mediaMoreOptions: 'Další možnosti',
      mediaSettings: 'Nastavení médií',
      collapse: 'Sbalit nabídku',
      expand: 'Rozbalte nabídku',
      more: 'Více',
      media: 'Média',
      tools: 'Nástroje',
      text: 'Text',
      advanced: 'Pokročilé',
      mediaMounted: 'Připojeno',
      mediaLibrary: 'Knihovna',
      textToHost: 'Do hostitele',
      textFromHost: 'Z hostitele'
    },
    ion: {
      checking: 'Kontrola videopaměti před spuštěním streamu...',
      warn: 'Videopaměti je málo. Jeden restart serveru by ji vyčerpal. Až se vám to bude hodit, restartujte.',
      criticalTitle: 'Nedostatek videopaměti ke spuštění streamu',
      criticalBody:
        'Spuštění videa by vyčerpalo vyhrazenou paměť a zastavilo server. Všechny ostatní funkce fungují dál, včetně ovládání napájení a restartu. Tuto paměť uvolní jen restart IronKVM.',
      criticalContinue: 'Přesto spustit video',
      criticalReboot: 'Restartovat IronKVM',
      criticalRebooting: 'Restartování...'
    }
  }
};

export default cz;
