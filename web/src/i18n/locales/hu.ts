const hu = {
  translation: {
    head: {
      desktop: 'Távoli Asztal',
      login: 'Bejelentkezés',
      changePassword: 'Jelszó megváltoztatása',
      terminal: 'Terminál',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        'A böngésző nem tárolta el a munkamenetet. Egy korábbi HTTPS-munkamenetből visszamaradt cookie titkosítatlan http-kapcsolaton nem cserélhető le. Törölje a cookie-kat ehhez a címhez, vagy nyisson privát ablakot, és jelentkezzen be újra.',
      login: 'Bejelentkezés',
      placeholderUsername: 'Adja meg a felhasználónevet',
      placeholderPassword: 'Adja meg a jelszót',
      placeholderCurrentPassword: 'Jelenlegi jelszó',
      placeholderPassword2: 'Adja meg újra a jelszót',
      noEmptyUsername: 'A felhasználónév nem lehet üres',
      noEmptyPassword: 'A jelszó nem lehet üres',
      passwordLength: 'A jelszónak 8 és 72 karakter között kell lennie',
      noAccount:
        'Nem sikerült megszerezni a felhasználói információkat, frissítse az oldalt vagy állítsa vissza a jelszót',
      invalidUser: 'Érvénytelen felhasználónév vagy jelszó',
      locked: 'Túl sok bejelentkezés, kérjük, próbálja újra később',
      globalLocked: 'A rendszer védelem alatt áll, próbálkozzon újra később',
      error: 'Váratlan hiba',
      invalidCurrentPassword: 'A jelenlegi jelszó helytelen',
      changePassword: 'Jelszó megváltoztatása',
      changePasswordDesc:
        'Az eszköz biztonsága érdekében módosítsa a webes bejelentkezési jelszót.',
      differentPassword: 'A jelszavak nem egyeznek',
      illegalUsername: 'A felhasználónév illegális karaktereket tartalmaz',
      illegalPassword: 'A jelszó illegális karaktereket tartalmaz',
      forgetPassword: 'Jelszó-emlékeztető',
      ok: 'Ok',
      cancel: 'Mégse',
      loginButtonText: 'Bejelentkezés',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the NanoKVM for 10 seconds.',
        reset2: 'A részletes lépésekért tekintse meg ezt a dokumentumot:',
        reset3: 'Alapértelmezett webes fiók:',
        reset4: 'Alapértelmezett SSH-fiók:',
        change1: 'Vegye figyelembe, hogy ez a művelet a következő jelszavakat módosítja:',
        change2: 'Webes bejelentkezési jelszó',
        change3: 'Rendszer root jelszava (SSH bejelentkezési jelszó)',
        change4: 'A jelszavak visszaállításához tartsa lenyomva a BOOT gombot a NanoKVM-en.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Wi-Fi beállítása a NanoKVM-hez',
      success: 'Please check the network status of NanoKVM and visit the new IP address.',
      failed: 'A művelet sikertelen, próbálja újra.',
      invalidMode:
        'Az aktuális mód nem támogatja a hálózat beállítását. Kérjük, lépjen az eszközére, és engedélyezze a Wi-Fi konfigurációs módot.',
      confirmBtn: 'Ok',
      finishBtn: 'Kész',
      ap: {
        authTitle: 'Hitelesítés szükséges',
        authDescription: 'A folytatáshoz adja meg a AP jelszót',
        authFailed: 'Érvénytelen AP jelszó',
        passPlaceholder: 'AP jelszót',
        verifyBtn: 'Ellenőrizze'
      }
    },
    screen: {
      scale: 'Skála',
      title: 'Képernyő',
      video: 'Videó mód',
      videoDirectTips:
        'Engedélyezze az HTTPS elemet a "Beállítások > Eszköz" menüpontban ennek a módnak a használatához',
      resolution: 'Felbontás',
      controlRegion: {
        title: 'Egérkalibrálás',
        description:
          'Akkor használja ezt a beállítást, ha a vezérelt eszköz nem 16:9 képarányú felbontást használ, és a kurzor vízszintesen vagy függőlegesen eltolódik.',
        off: 'Kikapcsolva',
        auto: 'Automatikus',
        autoWarning:
          'A kalibrálás sikertelen lehet, ha a felhasználói alkalmazás háttere teljesen fekete.',
        manual: 'Kézi',
        selectedResolution: 'Kijelölt terület felbontása',
        unused: 'Nincs használatban',
        originalResolution: 'Eredeti felbontás',
        selectResolution: 'Válassza ki az eredeti felbontást',
        addResolution: 'Egyéni felbontás hozzáadása',
        add: 'Hozzáadás',
        duplicateResolution: 'Ez a felbontás már létezik.',
        width: 'Szélesség',
        height: 'Magasság',
        apply: 'Számítás és alkalmazás',
        invalidResolution: 'A videó betöltése után adjon meg érvényes eredeti felbontást.',
        select: 'Terület kijelölése',
        clear: 'Automatikus felismerés visszaállítása',
        saveFailed: 'Nem sikerült menteni a bemeneti területet.',
        tooSmall: 'A kijelölt terület túl kicsi.',
        previewUnavailable: 'Az előnézet nem érhető el',
        clearConfirm: 'Visszaállítja a fekete szegélyek automatikus felismerését?',
        dragHint: 'Húzással jelölje ki a távoli asztal területét',
        finish: 'Kész',
        confirm: 'Megerősítés',
        cancel: 'Mégse'
      },
      auto: 'Automatikus',
      autoTips:
        'Bizonyos felbontások esetén képernyőszakadás vagy egéreltolódás léphet fel. Fontolja meg a távoli gép felbontásának módosítását vagy az automatikus mód kikapcsolását.',
      fps: 'FPS',
      customizeFps: 'Testreszabás',
      quality: 'Minőség',
      qualityLossless: 'Veszteségmentes',
      qualityHigh: 'Magas',
      qualityMedium: 'Közepes',
      qualityLow: 'Alacsony',
      frameDetect: 'Képkocka-figyelés',
      frameDetectTip:
        'Elemzi a képkockák közötti különbségeket. A videó stream küldése leáll, ha a távoli gép képernyőjén nem történik változás.',
      resetHdmi: 'HDMI visszaállítása',
      mixedH264: {
        title: 'H.264 adatfolyam-ütközés',
        description:
          'Az H.264 Direct és az H.264 WebRTC egyszerre van használatban. Ez képtörést vagy sérült videót okozhat. Csak egy H.264 módot használjon.'
      },
      webrtcConnectionFailed: {
        title: 'A WebRTC-kapcsolat sikertelen',
        description: 'Ellenőrizze a hálózati kapcsolatot, vagy váltson videómódot.'
      },
      captureStatus: {
        hdmiError: 'HDMI-képernyőhiba',
        unsupportedResolution: 'A jelenlegi felbontás nem támogatott',
        retrieving: 'Kép lekérése...',
        changingResolution: 'Felbontás váltása...',
        updateFailed: 'A kép jelenleg nem frissíthető',
        videoError: 'Videómegjelenítési hiba',
        noHdmi: 'Nem észlelhető HDMI-jel',
        unavailable: 'A kép jelenleg nem jeleníthető meg'
      }
    },
    keyboard: {
      title: 'Billentyűzet',
      paste: 'Beillesztés',
      tips: 'A szöveget billentyűleütésekként gépeli be a gazdagépen. Válassza ki a gazdagép billentyűzetkiosztását.',
      placeholder: 'Írja be',
      submit: 'Elküldés',
      virtual: 'Billentyűzet',
      readClipboard: 'Olvasás a vágólapról',
      clipboardPermissionDenied:
        'A vágólap engedélye megtagadva. Kérjük, engedélyezze a vágólaphoz való hozzáférést a böngészőjében.',
      clipboardReadError: 'Nem sikerült beolvasni a vágólapot',
      mediaKeys: {
        title: 'Médiabillentyűk',
        mute: 'Némítás',
        volumeDown: 'Hangerő le',
        volumeUp: 'Hangerő fel',
        previous: 'Előző szám',
        playPause: 'Lejátszás vagy szünet',
        next: 'Következő szám',
        stop: 'Leállítás'
      },
      pasting: {
        layout: 'A gazdagép billentyűzetkiosztása',
        layouts: {
          us: 'Angol (USA)',
          uk: 'Angol (Egyesült Királyság)',
          de: 'Német',
          fr: 'Francia',
          es: 'Spanyol',
          it: 'Olasz',
          ptBr: 'Portugál (Brazília)',
          se: 'Svéd / finn',
          ru: 'Orosz',
          ja: 'Japán',
          ko: 'Koreai'
        },
        speed: 'Gépelési sebesség',
        speeds: {
          fast: 'Gyors',
          normal: 'Normál',
          slow: 'Lassú'
        },
        estimate: 'Gépelési idő: kb. {{duration}}',
        untypeable: 'Karakterek, amelyeket ez a kiosztás nem tud begépelni: {{count}}',
        untypeableAt: '{{line}}. sor, {{column}}. oszlop',
        skipUntypeable: 'A többi begépelése',
        shortcut: 'A {{shortcut}} azonnal begépeli a vágólap tartalmát a gazdagépen.',
        clipboardUnavailable:
          'A böngésző csak HTTPS-en engedi, hogy egy oldal olvassa a vágólapot. Illessze be a szöveget a mezőbe a Ctrl+V billentyűkkel.',
        clipboardEmpty: 'A vágólap nem tartalmaz szöveget.',
        tooLong: 'A szöveg túl hosszú. A korlát {{max}} karakter.',
        inProgress: 'Már folyamatban van egy beillesztés begépelése.',
        typing: 'Gépelés a gazdagépen',
        done: 'Szöveg begépelve',
        canceled: 'Beillesztés megszakítva',
        failed: 'A beillesztés nem sikerült',
        cancel: 'Mégse',
        controlBusy: 'Egy másik vezérlő használja a billentyűzetet.',
        hidError: 'A billentyűleütéseket nem sikerült elküldeni a gazdagépnek.'
      },
      shortcut: {
        title: 'Parancsikonok',
        custom: 'Egyedi',
        capture: 'Kattintson ide a parancsikon rögzítéséhez',
        clear: 'Tiszta',
        save: 'Mentés',
        captureTips:
          'A rendszerszintű billentyűk (például a Windows billentyű) rögzítéséhez teljes képernyős engedély szükséges.',
        enterFullScreen: 'Teljes képernyős mód váltása.'
      },
      leaderKey: {
        title: 'Leader billentyű',
        desc: 'Kerülje ki a böngésző korlátozásait, és küldje el a rendszer parancsikonjait közvetlenül a távoli gazdagépnek.',
        howToUse: 'Használat',
        simultaneous: {
          title: 'Egyidejű üzemmód',
          desc1: 'Tartsa lenyomva a Leader billentyűt, majd nyomja meg a gyorsbillentyűt.',
          desc2: 'Intuitív, de ütközhet a rendszer parancsikonjaival.'
        },
        sequential: {
          title: 'Szekvenciális mód',
          desc1:
            'Nyomja meg a Leader billentyűt → nyomja meg sorban a gyorsbillentyűt → nyomja meg újra a Leader billentyűt.',
          desc2: 'Több lépést igényel, de teljesen elkerülhető a rendszerütközések.'
        },
        enable: 'Leader billentyű engedélyezése',
        tip: 'Leader billentyűként beállítva ez a billentyű kizárólag gyorsbillentyű-indítóként működik, és elveszíti alapértelmezett viselkedését.',
        placeholder: 'Nyomja meg a Leader billentyűt',
        shiftRight: 'Jobb Shift',
        ctrlRight: 'Jobb Ctrl',
        metaRight: 'Jobb Win',
        submit: 'Elküldés',
        recorder: {
          rec: 'REC',
          activate: 'Billentyűk aktiválása',
          input: 'Kérjük, nyomja meg a parancsikont...'
        }
      }
    },
    mouse: {
      title: 'Egér',
      cursor: 'Kurzorstílus',
      default: 'Alapértelmezett kurzor',
      pointer: 'Mutató kurzor',
      cell: 'Cella kurzor',
      text: 'Szöveg kurzor',
      grab: 'Markoló kurzor',
      hide: 'Kurzor elrejtése',
      mode: 'Egér mód',
      absolute: 'Abszolút mód',
      relative: 'Relatív mód',
      absoluteShort: 'Abszolút',
      relativeShort: 'Relatív',
      absoluteStalled: 'A célgép figyelmen kívül hagyja az abszolút egeret',
      absoluteStalledDesc:
        'A célgép már nem fogadja az abszolút egér jelentéseit, így a mutató mozgásai elvesznek. A billentyűzetet ez nem érinti. Az USB helyreállítása gyakran megoldja; a relatív mód másik végpontot használ.',
      useRelative: 'Váltás relatív módra',
      direction: 'Görgő iránya',
      scrollUp: 'Görgessen felfelé',
      scrollDown: 'Görgessen le',
      speed: 'Görgő sebessége',
      fast: 'Gyors',
      slow: 'Lassú',
      requestPointer:
        'Relatív mód használata. Kattintson az asztalra, hogy megjelenjen az egérmutató.',
      resetHid: 'HID alaphelyzetbe állítása',
      hidOnly: {
        title: 'Csak HID mód',
        desc: 'Ha az egér és a billentyűzet nem válaszol, és az HID alaphelyzetbe állítása nem segít, akkor az NanoKVM és az eszköz közötti kompatibilitási probléma lehet. Próbálja engedélyezni az HID-Csak módot a jobb kompatibilitás érdekében.',
        tip1: 'Az HID-Csak mód engedélyezése leválasztja a virtuális U-lemezt és a virtuális hálózatot',
        tip2: 'HID-Csak módban a képrögzítés le van tiltva',
        rebuild: 'A módváltás újraépíti az USB-kapcsolatot. A NanoKVM nem indul újra',
        enable: 'Engedélyezze a HID-Csak módot',
        disable: 'A HID-Csak mód letiltása'
      }
    },
    image: {
      title: 'Képek',
      loading: 'Betöltés...',
      empty: 'Nem található semmi',
      mountMode: 'Felszerelési mód',
      mountFailed: 'Csatlakoztatás sikertelen',
      mountDesc:
        'Egyes rendszerekben szükséges lehet a virtuális lemez eltávolítása a távoli gépen, mielőtt a képet csatlakoztatja.',
      unmountFailed: 'A leválasztás nem sikerült',
      unmountDesc:
        'Egyes rendszereken manuálisan kell kiadnia a távoli gazdagépről a kép leválasztása előtt.',
      refresh: 'Frissítse a képlistát',
      disk: 'Lemez',
      cdrom: 'CD',
      driveEmpty: 'Üres',
      eject: 'Kiadás',
      readOnly: 'Csak olvasható',
      readOnlyTip: 'A lemezbe következőként behelyezett képfájlra vonatkozik.',
      noDrives: 'Nincsenek virtuális meghajtók. Kapcsolja be a virtuális lemezt a Beállításokban.',
      insertFailed: 'A behelyezés sikertelen',
      ejectFailed: 'A kiadás sikertelen',
      insertInto: 'Behelyezés ide: {{drive}}. Kattintson a módosításhoz.',
      loadedIn: 'Ebben a meghajtóban: {{drive}}',
      attention: 'Figyelem',
      deleteConfirm: 'Biztosan törli ezt a képet?',
      okBtn: 'Igen',
      cancelBtn: 'Nem',
      tips: {
        title: 'Hogyan tölts fel képeket',
        usb1: 'Csatlakoztassa a NanoKVM-t a számítógépéhez USB-n keresztül.',
        usb2: 'Győződjön meg róla, hogy a virtuális lemez csatlakoztatva van (Beállítások - Virtuális lemez).',
        usb3: 'Nyissa meg a virtuális lemezt a számítógépén, és másolja a kép fájlt a virtuális lemez gyökérkönyvtárába.',
        scp1: 'Győződjön meg róla, hogy a NanoKVM és a számítógépe ugyanazon a helyi hálózaton van.',
        scp2: 'Nyisson meg egy terminált a számítógépén, és használja az SCP parancsot a kép fájl feltöltésére a /data könyvtárba a NanoKVM-en.',
        scp3: 'Példa: scp your-image-path root@your-nanokvm-ip:/data',
        tfCard: 'TF Kártya',
        tf1: 'Ez a módszer támogatott Linux rendszeren',
        tf2: 'Vegye ki a TF kártyát a NanoKVM-ből (a TELJES verzióhoz, először szedje szét a házat).',
        tf3: 'Helyezze a TF kártyát egy kártyaolvasóba, és csatlakoztassa a számítógépéhez.',
        tf4: 'Másolja a képfájlt a TF kártya /data könyvtárába.',
        tf5: 'Helyezze vissza a TF kártyát a NanoKVM-be.'
      }
    },
    script: {
      title: 'Szkriptek',
      upload: 'Feltöltés',
      run: 'Futtatás',
      runBackground: 'Háttérben futtatás',
      runFailed: 'Futtatás sikertelen',
      attention: 'Figyelem',
      delDesc: 'Biztosan törli ezt a fájlt?',
      confirm: 'Igen',
      cancel: 'Nem',
      delete: 'Törlés',
      close: 'Bezárás'
    },
    terminal: {
      title: 'Terminál',
      nanokvm: 'NanoKVM Terminál',
      serial: 'Soros port terminál',
      serialPort: 'Soros port',
      serialPortPlaceholder: 'Adja meg a soros portot',
      baudrate: 'Baudráta',
      parity: 'Paritás',
      parityNone: 'Nincs',
      parityEven: 'Páros',
      parityOdd: 'Páratlan',
      flowControl: 'Áramlásszabályozás',
      flowControlNone: 'Nincs',
      flowControlSoft: 'Szoftveres',
      flowControlHard: 'Hardveres',
      dataBits: 'Adatbitek',
      stopBits: 'Stop bitek',
      confirm: 'Ok'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'Parancs küldése...',
      sent: 'Parancs elküldve',
      input: 'Adja meg a MAC címet',
      ok: 'Ok'
    },
    download: {
      title: 'Képletöltő',
      input: 'Adjon meg egy távoli képet URL',
      ok: 'Ok',
      disabled: '/data partíció RO, ezért nem tudjuk letölteni a képet',
      uploadbox: 'Dobja ide a fájlt, vagy kattintson a kiválasztáshoz',
      inputfile: 'Kérjük, írja be a képfájlt',
      NoISO: 'Nincs ISO',
      sha256: 'SHA-256 (opcionális)',
      sha256Placeholder: 'Adjon meg egy 64 karakteres SHA-256 ellenőrzőösszeget',
      invalidSHA256: 'A SHA-256 értékének 64 karakteres hexadecimális karakterláncnak kell lennie',
      failed: 'Sikertelen letöltés',
      success: 'Sikeres letöltés',
      checksumFailed: 'Sikertelen letöltés: a SHA-256 ellenőrzése sikertelen',
      cancel: 'Mégse',
      cancelFailed: 'A letöltés megszakítása sikertelen',
      bootMenu: 'Rendszerindító menü (netboot.xyz)',
      bootMenuDesc: 'A netboot.xyz ISO letöltése ellenőrzött ellenőrzőösszeggel a virtuális CD-hez'
    },
    power: {
      title: 'Bekapcsolás',
      showConfirm: 'Megerősítés',
      showConfirmTip: 'Az áramellátási műveletekhez külön megerősítés szükséges',
      reset: 'Újraindítás',
      power: 'Bekapcsolás',
      powerShort: 'Bekapcsolás (rövid kattintás)',
      powerLong: 'Bekapcsolás (hosszú kattintás)',
      resetConfirm: 'Folytatja a visszaállítási műveletet?',
      powerConfirm: 'Folytatja az áramellátást?',
      okBtn: 'Igen',
      cancelBtn: 'Nem',
      hostOs: 'Gazdagép OS',
      hostOsTip:
        'USB-billentyűként kerülnek elküldésre. Hogy mit tesznek, azt a gazdagép dönti el.',
      sleep: 'Alvó állapot',
      wake: 'Ébresztés',
      wakeKey: 'Ébresztés Shifttel',
      powerDown: 'Leállítás',
      sleepConfirm: 'Alvó állapotba helyezi a gazdagépet?',
      powerDownConfirm: 'Elküldi a kikapcsoló billentyűt a gazdagépnek?',
      wakeTip:
        'Az alvó gazdagép gyakran figyelmen kívül hagyja az Ébresztést attól az eszköztől, amely elaltatta. Az Ébresztés Shifttel egy billentyűt nyom le a billentyűzeten, amelyet több gazdagép elfogad.',
      led: 'Bekapcsolásjelző LED',
      ledOn: 'Világít',
      ledOff: 'Nem világít',
      ledUnknown: 'Ismeretlen',
      ledConnected: 'Bekapcsolásjelző LED csatlakoztatva',
      ledConnectedTip:
        'Csak akkor kapcsolja be, ha a gazdagép bekapcsolásjelző LED-csatlakozója be van kötve a kártyára. Enélkül a tápellátás állapota ismeretlen.',
      ledConnectedFailed: 'Nem sikerült menteni a bekapcsolásjelző LED beállítását'
    },
    settings: {
      title: 'Beállítások',
      mcp: {
        title: 'MCP-szolgáltatás',
        service: 'MCP távoli vezérlés',
        serviceDesc:
          'Megbízható MCP-kliensek számára a billentyűzet és az egér vezérlésének, valamint képernyőképek készítésének engedélyezése',
        securityWarning:
          'Az API-kulcs birtokában bárki vezérelheti a távoli gazdagépet és láthatja annak képernyőjét. Használjon HTTPS-t, és csak megbízható hálózatokon engedélyezze.',
        endpoint: 'Végpont',
        apiKey: 'API-kulcs',
        regenerateConfirmTitle: 'Újragenerálja az MCP API-kulcsot?',
        regenerateConfirmDesc: 'A jelenlegi kulcs azonnal érvényét veszti.',
        enableConfirmTitle: 'Engedélyezi a külső MCP-vezérlést?',
        enableConfirmDesc:
          'Az MCP engedélyezése leállítja a PicoClaw-t, és bezár minden aktív PicoClaw-munkamenetet.',
        failed: 'Az MCP-művelet sikertelen',
        copyFailed: 'A másolás sikertelen. Másolja kézzel.',
        okBtn: 'Megerősítés',
        cancelBtn: 'Mégse'
      },
      redfish: {
        title: 'Redfish',
        service: 'Redfish szolgáltatás',
        serviceDesc:
          'A DMTF Redfish API tápellátás-vezérléshez, virtuális adathordozókhoz és állapotlekérdezéshez olyan eszközökből, mint a redfishtool és az Ansible. Kikapcsolásakor minden Redfish-munkamenet megszűnik.',
        endpoint: 'Szolgáltatás gyökere',
        httpsOn: 'A kártya HTTPS-t szolgál ki, amelyre a legtöbb Redfish-eszköznek szüksége van.',
        httpsOff:
          'A kártya titkosítatlan HTTP-t szolgál ki. A legtöbb Redfish-eszköznek HTTPS kell: kapcsolja be a "Beállítások > Hálózat" alatt.',
        credentials:
          'A Redfish a KVM-fiókokat fogadja el Basic hitelesítéssel vagy Redfish-munkamenettel, valamint X-Auth-Token fejlécben küldött API-kulcsokat. Az API-kulcsokat az API-kulcsok oldalon kezelheti.',
        powerActions: 'Tápellátási műveletek',
        powerActionsDesc:
          'A jelenleg elérhető reset típusok. Az On, a ForceOff és a GracefulShutdown művelethez ismerni kell a tápellátás állapotát, ezért ezek csak akkor érhetők el, ha a tápellátás menüben be van kapcsolva a "Bekapcsolásjelző LED csatlakoztatva".',
        sessions: 'Munkamenetek',
        noSessions: 'Nincs nyitott Redfish-munkamenet',
        created: 'Létrehozva',
        lastUsed: 'Utoljára használva',
        refresh: 'Frissítés',
        end: 'Befejezés',
        endConfirmTitle: 'Befejezi ezt a Redfish-munkamenetet?',
        endConfirmDesc: 'A tokenje azonnal érvényét veszti. A kliensnek újra be kell jelentkeznie.',
        failed: 'A Redfish-művelet sikertelen',
        copyFailed: 'A másolás sikertelen. Másolja kézzel.',
        okBtn: 'Megerősítés',
        cancelBtn: 'Mégse'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Gazdagép-watchdog',
        serviceDesc:
          'Ha a gazdagépnek működnie kellene, de a képe az időkorlát alatt nem változik, vagy nincs HDMI-jel, a panel megnyomja a resetet, vagy ki- és bekapcsolja a gazdagépet.',
        stillWarning:
          'Az a gazdagép, amelynek kijelzője alvó állapotba kerül, vagy amelynek képe munka közben mozdulatlan, lefagyottnak tűnik. Kapcsolja ki a kijelző alvását a gazdagépen, vagy adjon meg ping-címet.',
        ledHint:
          'A "Bekapcsolásjelző LED csatlakoztatva" kapcsoló ki van kapcsolva a tápellátás menüben. A watchdog nem látja, mikor van kikapcsolva a gazdagép, ezért mindig bekapcsoltnak tekinti.',
        timeout: 'Időkorlát',
        timeoutDesc: 'Mennyi ideig nem mutathat életjelet a gazdagép, mielőtt a watchdog közbelép.',
        action: 'Művelet',
        actionDesc:
          'A ki- és bekapcsolás 5 másodpercig nyomva tartja a bekapcsológombot, majd újra megnyomja.',
        actionReset: 'Reset',
        actionPower: 'Ki- és bekapcsolás',
        cooldown: 'Várakozási idő',
        cooldownDesc: 'A legrövidebb idő két művelet között.',
        maxPerHour: 'Művelet óránként',
        maxPerHourDesc: 'A műveletek legnagyobb száma egy órában.',
        pingHost: 'Ping-cím',
        pingHostDesc:
          'A gazdagép IP-címe. A válasz életjelnek számít. Hagyja üresen, ha nem kér pinget.',
        pingHostInvalid: 'Adjon meg egy IPv4- vagy IPv6-címet.',
        minutes: 'perc',
        save: 'Mentés',
        saved: 'Mentve',
        state: 'Érzékelő',
        status: {
          off: 'Kikapcsolva',
          watching: 'Figyel',
          hostOff: 'Gazdagép kikapcsolva',
          captureOff: 'HDMI-rögzítés kikapcsolva',
          cooldown: 'Várakozik',
          capped: 'Óránkénti korlát elérve',
          acting: 'Közbelép'
        },
        signal: 'HDMI-jel',
        yes: 'Igen',
        no: 'Nem',
        led: 'Tápellátás LED',
        on: 'Világít',
        off: 'Nem világít',
        ledNotConnected: 'Nincs csatlakoztatva',
        ping: 'Ping',
        pingNotSet: 'Nincs beállítva',
        pingReply: 'Válaszol',
        pingNoReply: 'Nincs válasz',
        lastChange: 'Utolsó képváltozás',
        never: 'Soha',
        actsIn: 'Közbelép ennyi idő múlva',
        actionsLastHour: 'Műveletek az elmúlt órában',
        duration: '{{minutes}} perc {{seconds}} mp',
        log: 'Napló',
        noLog: 'A watchdog még nem lépett közbe.',
        refresh: 'Frissítés',
        reasonFrozen: 'A kép nem változott',
        reasonNoSignal: 'Nincs HDMI-jel',
        stuckFor: '{{duration}} óta nincs életjel',
        pressFailed: 'A gombnyomás nem sikerült: {{error}}',
        noScreenshot: 'Nincs képernyőkép',
        failed: 'A watchdog művelete nem sikerült'
      },
      netboot: {
        title: 'Hálózati rendszerindítás',
        description:
          'A gazdagép indítása a hálózatról: iPXE és a KVM-en lévő lemezképek menüje az USB hálózati kapcsolaton, vagy netboot.xyz proxy DHCP-vel a LAN-on.',
        addon: 'dnsmasq és rendszerindító fájlok',
        addonDesc:
          'A /data-ra telepítve: a dnsmasq az Alpine-ból, az iPXE és a netboot.xyz a kiadásaikból, mindegyik az ellenőrzőösszegével ellenőrizve.',
        install: 'Telepítés',
        installing: 'Telepítés folyamatban. Ez néhány percig is tarthat.',
        uninstall: 'Eltávolítás',
        uninstallConfirm:
          'Kikapcsolja a hálózati rendszerindítást, és eltávolítja a dnsmasq-ot és a rendszerindító fájlokat?',
        needsData:
          'A hálózati rendszerindításhoz olyan IronKVM-lemezkép kell, amelyen a /data partíció csatolva van.',
        usb: 'Az USB hálózati kapcsolaton',
        usbDesc:
          'Amíg az USB hálózati kapcsolat be van kapcsolva, a dnsmasq szolgálja ki az udhcpd helyett. A gazdagép megkapja egyetlen címét útválasztó és DNS-kiszolgáló nélkül, az architektúrájának megfelelő iPXE-t és a KVM-en lévő ISO-lemezképek menüjét.',
        linkOff:
          'Az USB hálózati kapcsolat ki van kapcsolva. Kapcsolja be az Eszköz, USB-hálózat alatt.',
        menuUrl: 'Menü',
        leases: 'A gazdagép bérlete',
        noLeases: 'Még nincs',
        netbootxyzNote:
          'A menüben lévő netboot.xyz az internetről töltődik be, amelyet az USB-kapcsolat nem ér el. Ehhez a gazdagépnek egy másik hálózati porton kell internetet elérnie.',
        lan: 'Proxy DHCP a LAN-on',
        lanDesc:
          'A LAN PXE-klienseinek a netboot.xyz-t kínálja, amely ezután az internetről tölti be a menüjét. Soha nem oszt ki címet, és nem szolgálja ki a KVM-en lévő lemezképeket.',
        lanWarning:
          'A netboot.xyz-t ezen a LAN-on minden PXE-kliens megkapja, nem csak a gazdagép. Csak olyan hálózaton kapcsolja be, amelyet Ön felügyel.',
        lanConfirm: 'Bekapcsolja a proxy DHCP-t a LAN-on?',
        lanInterface: 'LAN',
        running: 'Fut',
        stopped: 'Nem fut',
        images: 'Lemezképek a menüben',
        noImages: 'Nincs ISO-lemezkép a lemezképkönyvtárban.',
        boots: 'Legutóbbi indítások',
        noBoots: 'A gazdagép még semmit sem töltött le.',
        log: 'dnsmasq-napló',
        refresh: 'Frissítés',
        okBtn: 'Megerősítés',
        cancelBtn: 'Mégse',
        failed: 'A hálózati rendszerindítási művelet sikertelen'
      },
      about: {
        title: 'NanoKVM Névjegy',
        information: 'Információ',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Alkalmazás verzió',
        applicationTip: 'NanoKVM webalkalmazás verziója',
        image: 'Képfájl verzió',
        imageTip: 'NanoKVM rendszerkép verziója',
        kernel: 'Kernelverzió',
        kernelTip: 'A jelenleg futó Linux-kernel kiadása',
        deviceKey: 'Eszköz kulcs',
        videoMemory: 'Videomemória',
        videoMemoryTip:
          'Videorögzítésre fenntartott memória. A rendszer többi része nem használja.',
        videoMemoryGenerations_one: '{{count}} korábbi NanoKVM-munkamenet foglal videomemóriát',
        videoMemoryGenerations_other: '{{count}} korábbi NanoKVM-munkamenet foglal videomemóriát',
        videoMemoryReboot: 'A felszabadításhoz indítsa újra.',
        community: 'Közösség',
        hostname: 'Gazdanév',
        hostnameUpdated: 'Gazdanév frissítve. Az alkalmazáshoz indítsa újra.',
        ipType: {
          Wired: 'Vezetékes',
          Wireless: 'Vezeték nélküli',
          Other: 'Egyéb'
        }
      },
      appearance: {
        title: 'Megjelenés',
        display: 'Kijelző',
        language: 'Nyelv',
        languageDesc: 'Válassza ki a felület nyelvét',
        webTitle: 'Webcím',
        webTitleDesc: 'A weboldal címének testreszabása',
        menuBar: {
          title: 'Menüsor',
          mode: 'Megjelenítési mód',
          modeDesc: 'Menüsor megjelenítése a képernyőn',
          modeOff: 'Ki',
          modeAuto: 'Automatikus elrejtés',
          modeAlways: 'Mindig látható',
          keyboardLedStatus: 'Billentyűzár-jelzők',
          keyboardLedStatusDesc:
            'A távoli számítógép Num Lock, Caps Lock és Scroll Lock állapotának megjelenítése',
          icons: 'Almenü ikonok',
          iconsDesc: 'Almenüikonok megjelenítése a menüsorban'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Távoli billentyűzárak állapota',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Be',
        off: 'Ki',
        unknown: 'Ismeretlen'
      },
      device: {
        title: 'Eszköz',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'OLED fényerő',
          brightnessDescription: 'Alacsonyabb szinten tovább bírja a kijelző',
          brightnessLevels: {
            '64': 'Legalacsonyabb',
            '96': 'Alacsony',
            '128': 'Közepes',
            '160': 'Magas',
            '207': 'Alapértelmezett',
            '255': 'Maximális'
          },
          0: 'Soha',
          15: '15 sec',
          30: '30 sec',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 óra'
        },
        ssh: {
          description: 'Engedélyezze a SSH távoli hozzáférést',
          tip: 'Az engedélyezés előtt állítson be erős jelszót (Fiók - Jelszó módosítása)'
        },
        advanced: 'Speciális beállítások',
        cpuFreq: {
          title: 'CPU-frekvencia',
          description: 'A következő rendszerindításkor alkalmazott CPU-órajel beállítása',
          tip: 'A CPU 850 MHz-en indul, és 1000 MHz-re van specifikálva. Az új érték a következő rendszerindításkor lép életbe, nem futás közben. Az 1000 MHz a specifikáción belül van; a hőmérséklet mindkét beállításnál jóval a határértékek alatt marad.',
          running: 'Jelenleg: {{mhz}} MHz',
          rebootToApply: 'az alkalmazáshoz újraindítás szükséges',
          rebootConfirm: 'Újraindítja most, hogy a {{mhz}} MHz érvénybe lépjen?'
        },
        swap: {
          title: 'Csere',
          disable: 'Letiltás',
          description: 'Állítsa be a swap fájl méretét',
          tip: 'Ennek a funkciónak az engedélyezése lerövidítheti az SD-kártya élettartamát!'
        },
        zram: {
          title: 'Tömörített swap (zram)',
          description: 'Swap tömörített RAM-ban az SD-kártya helyett',
          tip: 'A zram távol tartja a swapot az SD-kártyától, így nem koptatja azt. Mögötte nincs lemezes swap: ha a zram megtelik, a kernel leállít egy folyamatot a lassú lapozás helyett. A memóriakorlát határozza meg, mennyi RAM-ot foglalhat a zram.',
          unavailable: 'A kernelmodulok nincsenek telepítve ezen az eszközön',
          inactive: 'Engedélyezve, de az eszköz nem indult el',
          active: 'Aktív - {{used}} / {{total}}, {{ratio}}x',
          off: 'Kikapcsolva',
          detail: {
            algorithm: 'Algoritmus: {{algorithm}}',
            memory: 'Használt memória: {{used}} / {{limit}}',
            memoryNoLimit: 'Használt memória: {{used}}, nincs korlát beállítva',
            counters:
              'Beolvasott lapok: {{in}}, kiírt lapok: {{out}} (minden swap eszköz, rendszerindítás óta)'
          }
        },
        mouseJiggler: {
          title: 'Mouse Jiggler',
          description: 'A távoli gazdagép alvó állapotának megakadályozása',
          disable: 'Letiltás',
          absolute: 'Abszolút mód',
          relative: 'Relatív mód'
        },
        mdns: {
          description: 'Engedélyezze az mDNS felderítési szolgáltatást',
          tip: 'Kikapcsolás, ha nincs rá szükség'
        },
        hdmi: {
          description: 'HDMI/monitor kimenet engedélyezése',
          idleTimeoutTitle: 'Inaktív rögzítés időkorlátja',
          idleTimeoutDescription: 'A HDMI-rögzítés leállítása, ha nincs aktív néző ennyi ideig:',
          minutes: 'perc'
        },
        autostart: {
          title: 'Automatikus indítási parancsfájlok beállításai',
          description: 'A rendszer indításakor automatikusan futó szkriptek kezelése',
          new: 'Új',
          deleteConfirm: 'Biztosan törli ezt a fájlt?',
          yes: 'Igen',
          no: 'Nem',
          scriptName: 'Automatikusan induló szkript neve',
          scriptContent: 'A szkripttartalom automatikus indítása',
          settings: 'Beállítások'
        },
        hidOnly: 'HID-Csak mód',
        hidOnlyDesc:
          'A virtuális eszközök emulálásának leállítása, csak az alapvető HID vezérlés megtartásával',
        disk: 'Virtuális lemez',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Virtuális hálózat',
        networkDesc: 'Virtuális hálózati kártya csatlakoztatása a távoli gazdagépen',
        usbNetwork: {
          description:
            'Privát hálózati kapcsolat a távoli gazdagéppel az USB-kábelen keresztül. A gazdagép átjáró és DNS nélküli címet kap, így a NanoKVM-en keresztül nem éri el a helyi hálózatot.',
          off: 'Ki',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (NCM nélküli gazdagépekhez)',
          rndis: 'RNDIS (már nem választható)',
          rndisNote:
            'Ez a kapcsolat RNDIS-t használ, amely már nem választható. Válassza az NCM-et vagy az ECM-et.',
          subnet: 'Alhálózat',
          subnetDesc:
            'Privát IPv4-hálózat, /24 és /30 között. A NanoKVM az első címet kapja, a gazdagép a másodikat.',
          addresses: 'NanoKVM: {{board}}, gazdagép: {{host}}',
          invalidSubnet: 'Adjon meg egy alhálózatot, például 172.31.255.0/30.',
          apply: 'Alkalmaz',
          confirm: 'Újracsatlakoztatja az USB-eszközt?',
          reenumerate:
            'Az alkalmazás újraépíti az USB-kapcsolatot. A gazdagép néhány másodpercre elveszíti a billentyűzetet, az egeret és a virtuális lemezt.'
        },
        audio: 'Virtuális hangszóró',
        audioDesc:
          'USB-hangkártyát jelenít meg a távoli gazdagépen, így hallhatja annak hangját. A gazdagépen ki kell választani kimeneti eszközként. A kapcsolása újraépíti az USB-kapcsolatot.',
        audioNote: 'Hang mindkét H.264 módban (WebRTC és Direct) elérhető, MJPEG módban nem',
        console: 'Soros konzol',
        consoleDesc:
          'USB soros portot jelenít meg a távoli gazdagépen, amelyen át bejelentkezhet erre a NanoKVM-re, ha a hálózat nem érhető el',
        consoleTip:
          'Bárki, aki a távoli gazdagépet vezérli, bejelentkezési promptot kap ehhez a NanoKVM-hez. Az engedélyezés előtt állítson be erős jelszót (Fiók - Jelszó módosítása).',
        endpoints: {
          title: 'USB-végpontok',
          used: '{{used}} / {{total}} használatban',
          cost: '{{cost}} foglalt',
          needs: '{{cost}} szükséges',
          full: 'Nincs elég USB-végpont. Előbb kapcsoljon ki valami mást.',
          inactive:
            'Bekapcsolva, de nem fut: az USB-vezérlőnek elfogytak a végpontjai. Kapcsoljon ki egy másik eszközt, és ez azonnal elindul.',
          explain:
            'Az USB-vezérlőnek rögzített számú bemeneti végpontja van, ez a számláló ezeket mutatja. Ha több eszköz van engedélyezve, mint amennyi elfér, a billentyűzet és az egér megmarad, a többi kikapcsol.',
          error: 'Az eszköz nem érhető el. Próbálja újra.',
          fitTogether: 'Ezek együtt elférnek: {{sets}}'
        },
        reboot: 'Újraindítás',
        rebootDesc: 'Biztos, hogy újra akarja indítani a NanoKVM-t?',
        okBtn: 'Igen',
        cancelBtn: 'Nem'
      },
      network: {
        title: 'Hálózat',
        wifi: {
          title: 'Wi-Fi',
          description: 'Wi-Fi beállítása',
          apMode: 'Az AP mód engedélyezve van, csatlakozzon a Wi-Fihez a QR-kód beolvasásával',
          connect: 'Wi-Fi csatlakoztatása',
          connectDesc1: 'Adja meg a hálózat SSID-jét és jelszavát',
          connectDesc2: 'Adja meg a jelszót a hálózathoz való csatlakozáshoz',
          disconnect: 'Biztosan bontja a hálózati kapcsolatot?',
          failed: 'A csatlakozás sikertelen, próbálja újra.',
          ssid: 'Név',
          password: 'Jelszó',
          joinBtn: 'Csatlakozás',
          confirmBtn: 'OK',
          cancelBtn: 'Mégse'
        },
        tls: {
          description: 'HTTPS protokoll engedélyezése',
          tip: 'Figyelem: A HTTPS használata növelheti a késleltetést, különösen MJPEG videó módban.',
          restarting: 'Az eszköz szervere újraindul, ez körülbelül két percig tart...',
          waiting: 'Várakozás, hogy az eszköz újra válaszoljon...',
          waitingHttp: 'Visszaváltás http-re. Ha az oldal nem nyílik meg magától, töltse újra.'
        },
        ethernet: {
          title: 'IP-cím',
          description: 'Állítsa be, hogyan kapja a NanoKVM a címét a vezetékes hálózaton',
          dhcp: 'DHCP',
          manual: 'Kézi',
          networkDetails: 'Hálózati adatok',
          interface: 'Interfész',
          ipAddress: 'IP-cím',
          subnetMask: 'Alhálózati maszk',
          router: 'Útválasztó',
          save: 'Alkalmaz',
          invalidAddress: 'Adjon meg egy érvényes IP-címet',
          invalidMask: 'Adjon meg egy érvényes alhálózati maszkot, például 255.255.255.0 vagy 24',
          invalidRouter: 'Adjon meg egy érvényes útválasztócímet',
          addressRequired: 'IP-cím megadása kötelező',
          maskRequired: 'Alhálózati maszk megadása kötelező',
          applyTitle: 'Megváltoztatja a NanoKVM címét?',
          applyWarning:
            'A kapcsolat ezzel az oldallal megszakad. A NanoKVM alkalmazza az új címet, és {{seconds}} másodpercet vár arra, hogy elérje azon a címen. Az elérés megtartja a módosítást. Ha semmi sem éri el, a NanoKVM visszaállítja a korábbi beállításokat.',
          applyConfirm: 'Alkalmaz',
          applyCancel: 'Mégse',
          applyFailed: 'A cím alkalmazása nem sikerült',
          trialTitle: 'Várakozás a megerősítésre',
          trialDhcp: 'A NanoKVM címet kér a DHCP-től.',
          trialStatic: 'A NanoKVM most a következő címen érhető el: {{address}}.',
          trialInstruction:
            'Nyissa meg a NanoKVM-et az új címén, és jelentkezzen be, ha kéri. Az elérés megtartja a módosítást. Ha {{seconds}} másodpercen belül semmi sem éri el a NanoKVM-et, visszaállítja a korábbi beállításokat.',
          trialOpen: 'Az új cím megnyitása',
          trialKeep: 'Beállítások megtartása',
          trialKept: 'Az új cím mentve',
          trialKeepFailed: 'A beállításokat nem sikerült megtartani',
          trialGone: 'A módosítás már vissza lett vonva. Próbálja újra.',
          unsaved: 'Nem mentett módosítások'
        },
        dns: {
          title: 'DNS',
          description: 'DNS-kiszolgálók beállítása a NanoKVM számára',
          mode: 'Mód',
          dhcp: 'DHCP',
          manual: 'Kézi',
          add: 'DNS hozzáadása',
          save: 'Mentés',
          invalid: 'Adjon meg egy érvényes IP-címet',
          noDhcp: 'Jelenleg nincs elérhető DHCP DNS',
          saved: 'DNS-beállítások mentve',
          saveFailed: 'Nem sikerült menteni a DNS-beállításokat',
          unsaved: 'Nem mentett módosítások',
          maxServers: 'Legfeljebb {{count}} DNS-kiszolgáló engedélyezett',
          dnsServers: 'DNS-kiszolgálók',
          dhcpServersDescription: 'A DNS-kiszolgálók automatikusan DHCP-n keresztül érkeznek',
          manualServersDescription: 'A DNS-kiszolgálók kézzel szerkeszthetők',
          networkDetails: 'Hálózati részletek',
          interface: 'Interfész',
          ipAddress: 'IP-cím',
          subnetMask: 'Alhálózati maszk',
          router: 'Router',
          none: 'Nincs'
        }
      },
      vpn: {
        loading: 'Betöltés...',
        okBtn: 'Igen',
        cancelBtn: 'Nem',
        restart: '{{name}} újraindítása?',
        stop: '{{name}} leállítása?',
        stopDesc:
          'A démon most leáll. Az indításkori automatikus indítás külön kapcsoló, és változatlan marad.',
        update: '{{name}} frissítése erre: {{version}}?',
        updateDesc: 'Ha a démon fut, újraindul. A bejelentkezés megmarad.',
        notInstall: '{{name}} nincs telepítve.',
        install: 'Telepítés',
        installing: 'Telepítés folyamatban',
        installFailed: 'A telepítés sikertelen',
        retry: 'Újrapróbálás',
        notRunning: '{{name}} nem fut. A folytatáshoz indítsa el.',
        run: 'Indítás',
        boot: 'Indítás rendszerindításkor',
        bootDesc: '{{name}} indítása a KVM indulásakor.',
        enable: '{{name}} engedélyezése',
        control: 'Vezérlőszerver',
        connected: 'Csatlakozva',
        disconnected: 'Nincs csatlakozva',
        deviceName: 'Eszköznév',
        deviceIP: 'Eszköz IP-címe',
        account: 'Fiók',
        version: 'Verzió',
        uptime: 'Üzemidő',
        peers: 'Társak',
        noPeers: 'Még nincsenek társak.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Memória',
        daemonRss: 'Démon',
        group: 'Bővítménycsoport',
        high: '{{size}} felett lassítva',
        max: '{{size}} felett a kernel leállítja',
        noGroup: 'Ezen a kártyán nincs bővítmény-memóriacsoport.',
        uninstall: '{{name}} eltávolítása',
        uninstallDesc: 'Biztosan eltávolítja ezt: {{name}}? A bejelentkezés megmarad a kártyán.',
        blocked:
          '{{other}} fut, vagy rendszerindításkor elindul. Egyszerre csak egy VPN futhat: előbb állítsa le ezt: {{other}}, és kapcsolja ki az automatikus indítását.',
        swap: {
          title: 'Swap memória',
          tip: 'Ha a démonnak kevés a memóriája, próbálja engedélyezni a swap memóriát. Ez alapértelmezés szerint 256MB-ra állítja a swap fájl méretét, amely a "Beállítások > Eszköz" alatt módosítható.'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Frissítse az oldalt, majd próbálja újra. Vagy próbálja meg manuálisan telepíteni.',
        download: 'Letöltés a',
        package: 'telepítési csomag',
        unzip: 'és kicsomagolás',
        upTailscale: 'Töltsön fel tailscale-t a NanoKVM /usr/bin/ könyvtárába',
        upTailscaled: 'Töltsön fel tailscaled-t a NanoKVM /usr/sbin/ könyvtárába',
        refresh: 'Frissítse az aktuális oldalt',
        notLogin:
          'Az eszköz még nincs kötve. Kérem, jelentkezzen be és kösse az eszközt a fiókjához.',
        urlPeriod: 'Ez az url 10 percig érvényes',
        login: 'Bejelentkezés',
        loginSuccess: 'Sikeres bejelentkezés',
        logout: 'Kijelentkezés',
        logoutDesc: 'Biztos, hogy ki szeretne jelentkezni?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Ez az eszköz még nem csatlakozott NetBird-hálózathoz. Csatlakozzon beállítókulccsal, vagy jelentkezzen be SSO-val.',
        setupKey: 'Beállítókulcs',
        setupKeyPlaceholder: 'Illesszen be egy beállítókulcsot a NetBird irányítópultjáról',
        join: 'Csatlakozás',
        or: 'vagy',
        sso: 'Bejelentkezés SSO-val',
        urlPeriod: 'Ez az url 10 percig érvényes',
        loginSuccess: 'Sikeres bejelentkezés',
        logout: 'Regisztráció törlése',
        logoutDesc:
          'A regisztráció törlése eltávolítja ezt a társat a NetBird-fiókjából, és törli itt a konfigurációját. Az újbóli csatlakozáshoz beállítókulcs vagy SSO-bejelentkezés kell, és a társ új IP-címet kaphat. Folytatja?'
      },
      update: {
        title: 'Frissítés keresése',
        queryFailed: 'Verzió lekérdezése sikertelen',
        updateFailed: 'Frissítés sikertelen. Kérem, próbálja újra.',
        isLatest: 'Ön már a legfrissebb verziót használja.',
        available: 'Frissítés elérhető. Biztos, hogy frissít?',
        updating: 'Frissítés elkezdődött. Kérem várjon...',
        confirm: 'Megerősítés',
        cancel: 'Mégse',
        preview: 'Frissítések előnézete',
        previewDesc: 'Korai hozzáférést kap az új funkciókhoz és fejlesztésekhez',
        previewTip:
          'Kérjük, vegye figyelembe, hogy az előzetes verziók hibákat vagy hiányos funkciókat tartalmazhatnak!',
        customServer: {
          title: 'Egyéni frissítési kiszolgáló',
          desc: 'Online frissítések keresése és letöltése a megadott kiszolgálóról',
          invalidUrl:
            'Adjon meg egy érvényes HTTP- vagy HTTPS-kiszolgálókönyvtárat lekérdezés, töredékazonosító és latest.json nélkül.',
          loadFailed: 'Nem sikerült betölteni a frissítési kiszolgáló beállításait.',
          saveFailed: 'Nem sikerült menteni a frissítési kiszolgáló beállításait.',
          saved: 'A frissítési kiszolgáló beállításai mentve.',
          save: 'Mentés',
          confirmTitle: 'Egyéni frissítési kiszolgálót használ?',
          confirmDesc:
            'Az SHA-512 csak azt ellenőrzi, hogy a csomag megfelel-e a kiszolgáló által biztosított jegyzéknek. Nem igazolja, hogy a csomag hivatalos NanoKVM-kiadás. Egy hibás vagy rosszindulatú kiszolgáló használhatatlanná teheti az eszközt, adatvesztést okozhat, vagy veszélyeztetheti a rendszert.',
          confirm: 'Használat mindenképpen',
          useSipeed: 'A hivatalos Sipeed kiszolgáló használata',
          previewDisabled:
            'Az előzetes frissítések nem érhetők el, amíg egyéni frissítési kiszolgáló van engedélyezve.'
        },
        offline: {
          title: 'Offline frissítések',
          desc: 'Frissítés helyi telepítőcsomaggal',
          upload: 'Feltöltés',
          checksumPlaceholder: 'SHA-256 ellenőrzőösszeg (opcionális)',
          invalidChecksum: 'A SHA-256 ellenőrzőösszegnek 64 hexadecimális karakterből kell állnia.',
          checksumMismatch: 'Az SHA-256 ellenőrzése sikertelen. Lehet, hogy a csomag sérült.',
          invalidName: 'Érvénytelen fájlnévformátum. Kérjük, töltse le a GitHub kiadásaiból.',
          updateFailed: 'Frissítés sikertelen. Kérem, próbálja újra.'
        }
      },
      account: {
        title: 'Fiók',
        webAccount: 'Webes fiók neve',
        role: 'Szerepkör',
        roles: { admin: 'Rendszergazda', user: 'Felhasználó' },
        password: 'Jelszó',
        updateBtn: 'Update',
        logoutBtn: 'Kijelentkezés',
        logoutDesc: 'Biztos, hogy ki szeretne jelentkezni?',
        okBtn: 'Igen',
        cancelBtn: 'Nem',
        users: {
          title: 'Felhasználók',
          create: 'Felhasználó létrehozása',
          enabled: 'Engedélyezve',
          disabled: 'Letiltva',
          deviceOwner: 'Eszköz tulajdonosa',
          resetPassword: 'Jelszó visszaállítása',
          delete: 'Törlés',
          deleteConfirm: 'Törli ezt a felhasználót, és visszavonja az összes munkamenetét?',
          created: 'Felhasználó létrehozva',
          deleted: 'Felhasználó törölve',
          passwordUpdated: 'Jelszó frissítve',
          loadFailed: 'Nem sikerült betölteni a felhasználókat',
          saveFailed: 'Nem sikerült menteni a felhasználót',
          deleteFailed: 'Nem sikerült törölni a felhasználót'
        }
      },
      apiKeys: {
        title: 'API-kulcsok',
        description:
          'A kulcs a tulajdonosa nevében, annak szerepkörével működik. Küldje Authorization: Bearer <key> fejlécként a metrikákhoz és az API-hoz, vagy X-Auth-Token fejlécként a Redfish-hez.',
        name: 'Név',
        namePlaceholder: 'Mire szolgál a kulcs, például prometheus',
        nameRequired: 'Adjon nevet a kulcsnak',
        nameTooLong: 'A név legfeljebb 64 karakter lehet',
        unnamed: '(névtelen)',
        create: 'Kulcs létrehozása',
        created: 'Létrehozva',
        owner: 'Tulajdonos',
        empty: 'Nincsenek API-kulcsok',
        newKeyTitle: 'Az új API-kulcsa',
        newKeyWarning:
          'Másolja ki a kulcsot most. Nem kerül tárolásra, és később nem jeleníthető meg újra. Ha elveszíti, vonja vissza, és hozzon létre egy újat.',
        copy: 'Másolás',
        copied: 'Másolva',
        copyFailed: 'A másolás sikertelen. Másolja kézzel.',
        done: 'Kész',
        revoke: 'Visszavonás',
        revokeConfirmTitle: 'Visszavonja ezt az API-kulcsot?',
        revokeConfirmDesc: 'Minden, ami a(z) "{{name}}" kulcsot használja, azonnal leáll.',
        revoked: 'API-kulcs visszavonva',
        loadFailed: 'Nem sikerült betölteni az API-kulcsokat',
        createFailed: 'Nem sikerült létrehozni az API-kulcsot',
        revokeFailed: 'Nem sikerült visszavonni az API-kulcsot',
        cancelBtn: 'Mégse'
      }
    },
    picoclaw: {
      title: 'PicoClaw Asszisztens',
      empty: 'Nyissa meg a panelt, és indítsa el a feladatot.',
      inputPlaceholder: 'Írja le, mit szeretne tenni az PicoClaw-val',
      newConversation: 'Új beszélgetés',
      processing: 'Feldolgozás...',
      agent: {
        defaultTitle: 'Általános asszisztens',
        defaultDescription: 'Általános csevegési, keresési és munkaterületi súgó.',
        kvmTitle: 'Távoli vezérlés',
        kvmDescription: 'Működtesse a távoli gazdagépet az NanoKVM segítségével.',
        switched: 'Ügynöki szerepkör megváltozott',
        switchFailed: 'Nem sikerült váltani az ügynöki szerepkört'
      },
      send: 'Küldés',
      cancel: 'Mégse',
      status: {
        connecting: 'Csatlakozás az átjáróhoz...',
        connected: 'PicoClaw munkamenet csatlakoztatva',
        disconnected: 'PicoClaw munkamenet lezárva',
        stopped: 'Leállítási kérés elküldve',
        runtimeStarted: 'PicoClaw Runtime elindult',
        runtimeStartFailed: 'Nem sikerült elindítani a PicoClaw Runtime-ot',
        runtimeStopped: 'PicoClaw Runtime leállt',
        runtimeStopFailed: 'Nem sikerült leállítani a PicoClaw Runtime-ot',
        controlSwitchedToMCP: 'A vezérlés átkerült a külső MCP-szolgáltatáshoz'
      },
      connection: {
        runtime: {
          checking: 'Ellenőrzés',
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime kész',
          stopped: 'Runtime leállt',
          blockedByMCP: 'A külső MCP-vezérlés aktív',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime nem érhető el',
          configError: 'Konfigurációs hiba'
        },
        transport: {
          connecting: 'Csatlakozás',
          connected: 'Csatlakoztatva',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'Üresjárat',
          busy: 'Elfoglalt'
        }
      },
      message: {
        toolAction: 'Akció',
        observation: 'Megfigyelés',
        screenshot: 'Képernyőkép'
      },
      overlay: {
        locked: 'PicoClaw vezérli az eszközt. A kézi bevitel szünetel.'
      },
      control: {
        picoclaw: 'Eszközvezérlés: PicoClaw',
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Eszközvezérlés: külső MCP',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Eszközvezérlés: kikapcsolva',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Vezérlés átadása',
        release: 'Vezérlés feloldása',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'PicoClaw-vezérlés megadva',
        released: 'PicoClaw-vezérlés feloldva',
        grantFailed: 'Nem sikerült megadni a PicoClaw-vezérlést',
        releaseFailed: 'Nem sikerült feloldani a PicoClaw-vezérlést',
        grantConfirmTitle: 'Átváltja az eszközvezérlést PicoClaw-ra?',
        grantConfirmDesc: 'A külső MCP eszközírásai megszakadnak.'
      },
      install: {
        install: 'PicoClaw telepítése',
        installing: 'PicoClaw telepítése folyamatban',
        success: 'PicoClaw sikeresen telepítve',
        failed: 'Nem sikerült telepíteni PicoClaw',
        uninstalling: 'Runtime eltávolítása...',
        uninstalled: 'A Runtime sikeresen eltávolítva.',
        uninstallFailed: 'Az eltávolítás nem sikerült.',
        requiredTitle: 'PicoClaw nincs telepítve',
        requiredDescription:
          'Telepítse a PicoClaw alkalmazást a PicoClaw Runtime elindítása előtt.',
        progressDescription: 'PicoClaw letöltése és telepítése folyamatban van.',
        stages: {
          preparing: 'Felkészülés',
          downloading: 'Letöltés',
          extracting: 'Kibontás',
          verifying: 'Ellenőrzés',
          installing: 'Telepítés folyamatban',
          installed: 'Telepítve',
          install_timeout: 'Időtúllépés',
          install_failed: 'Sikertelen'
        }
      },
      model: {
        requiredTitle: 'Modellkonfiguráció szükséges',
        requiredDescription: 'A PicoClaw chat használata előtt konfigurálja az PicoClaw modellt.',
        docsTitle: 'Konfigurációs útmutató',
        docsDesc: 'Támogatott modellek és protokollok',
        menuLabel: 'Modell konfigurálása',
        modelIdentifier: 'Modellazonosító',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'API-kulcs',
        apiKeyPlaceholder: 'Adja meg a modell API-kulcsát',
        save: 'Mentés',
        saving: 'Mentés',
        saved: 'A modell konfigurációja mentve',
        saveFailed: 'Nem sikerült menteni a modellkonfigurációt',
        invalid: 'A modellazonosító, az API Base URL és az API-kulcs megadása kötelező'
      },
      uninstall: {
        menuLabel: 'Eltávolítás',
        confirmTitle: 'Eltávolítás PicoClaw',
        confirmContent:
          'Biztosan eltávolítja a következőt: PicoClaw? Ezzel törli a végrehajtható fájlt és az összes konfigurációs fájlt.',
        confirmOk: 'Eltávolítás',
        confirmCancel: 'Mégse'
      },
      history: {
        title: 'Előzmények',
        loading: 'Munkamenetek betöltése...',
        emptyTitle: 'Még nincs előzmény',
        emptyDescription: 'A korábbi PicoClaw munkamenetek itt jelennek meg.',
        loadFailed: 'Nem sikerült betölteni a munkamenet-előzményeket',
        deleteFailed: 'Nem sikerült törölni a munkamenetet',
        deleteConfirmTitle: 'Munkamenet törlése',
        deleteConfirmContent: 'Biztos, hogy törölni szeretné a következőt: "{{title}}"?',
        deleteConfirmOk: 'Törlés',
        deleteConfirmCancel: 'Mégse',
        messageCount_one: '{{count}} üzenet',
        messageCount_other: '{{count}} üzenet',
        messageCount: '{{count}} üzenet'
      },
      config: {
        startRuntime: 'PicoClaw indítása',
        stopRuntime: 'PicoClaw leállítása'
      },
      start: {
        enableConfirmTitle: 'Átváltja a vezérlést a PicoClaw-ra?',
        enableConfirmDesc: 'A PicoClaw indítása letiltja a külső MCP-szolgáltatást.',
        enableConfirmOk: 'PicoClaw indítása',
        enableConfirmCancel: 'Mégse',
        title: 'PicoClaw indítása',
        description: 'Indítsa el a Runtime-ot a PicoClaw segéd használatának megkezdéséhez.',
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'Problémába ütköztünk',
      refresh: 'Frissítés',
      panel: 'Az oldal ezen része nem működik',
      retry: 'Újra'
    },
    fullscreen: {
      toggle: 'Teljes képernyő váltás'
    },
    input: {
      disconnected: 'A billentyűzet és az egér nincs csatlakoztatva',
      disconnectedTls:
        'A böngésző kérdés nélkül elutasította a billentyűzetet és az egeret továbbító biztonságos kapcsolatot. Az eszköz által generált tanúsítvány még nem megbízható. Nyissa meg ezt a címet egy új lapon, fogadja el a tanúsítványt, majd töltse újra az oldalt. A megbízható megoldás a tanúsítvány telepítése.',
      disconnectedNever:
        'A billentyűzetet és az egeret továbbító kapcsolatot nem sikerült megnyitni. Az oldal többi része működik, mert nem használja ezt a kapcsolatot. Ellenőrizze, hogy semmi sem blokkolja Ön és az eszköz között.',
      disconnectedDropped:
        'A billentyűzetet és az egeret továbbító kapcsolat megszakadt, és nem állt helyre. Újraindítás után magától újracsatlakozik; ha ez az állapot megmarad, töltse újra az oldalt.',
      hidDisabled: 'A HID ki van kapcsolva ezen az eszközön (/boot/disable_hid).',
      keyFailed: 'A billentyűt nem sikerült elküldeni.'
    },
    speaker: { title: 'Hangszóró', unmute: 'Némítás feloldása', mute: 'Némítás' },
    menu: {
      collapse: 'Menü összecsukása',
      expand: 'Bontsa ki a menüt'
    },
    ion: {
      checking: 'Videomemória ellenőrzése az adatfolyam indítása előtt...',
      warn: 'Kevés a videomemória. Egyetlen szerver-újraindítás elfogyasztaná. Indítsa újra, amikor alkalmas.',
      criticalTitle: 'Nincs elég videomemória az adatfolyam indításához',
      criticalBody:
        'A videó indítása elfogyasztaná a fenntartott memóriát, és leállítaná a szervert. Minden más funkció továbbra is működik, beleértve a tápellátás-vezérlést és az újraindítást. Ezt a memóriát csak a NanoKVM újraindítása szabadítja fel.',
      criticalContinue: 'Videó indítása mégis',
      criticalReboot: 'NanoKVM újraindítása',
      criticalRebooting: 'Újraindítás...'
    }
  }
};

export default hu;
