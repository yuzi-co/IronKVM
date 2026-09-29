const hu = {
  translation: {
    feedback: {
      enabled: '{{name}} bekapcsolva',
      disabled: '{{name}} kikapcsolva',
      failed: 'A kérés sikertelen. Próbálja újra.',
      network: 'Az eszköz nem érhető el. Ellenőrizze a kapcsolatot, és próbálja újra.',
      saved: 'Mentve',
      timeout: 'Az eszköz túl sokáig nem válaszolt. Próbálja újra.'
    },
    common: {
      copy: 'Másolás',
      copied: 'Másolva',
      copyFailed: 'A másolás nem sikerült. Jelölje ki a szöveget, és másolja kézzel.',
      notUpdating: 'Nem frissül: az utolsó lekérdezés sikertelen volt.',
      off: 'Ki',
      running: 'Fut',
      save: 'Mentés',
      cancel: 'Mégse',
      delete: 'Törlés',
      remove: 'Eltávolítás'
    },
    head: {
      desktop: 'Távoli Asztal',
      login: 'Bejelentkezés',
      changePassword: 'Jelszó megváltoztatása',
      terminal: 'Terminál',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'A jelszó megváltozott. Jelentkezzen be az új jelszóval.',
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
          'To reset the passwords, pressing and holding the BOOT button on the IronKVM for 10 seconds.',
        reset3: 'Alapértelmezett webes fiók:',
        reset4: 'Alapértelmezett SSH-fiók:',
        change1: 'Vegye figyelembe, hogy ez a művelet a következő jelszavakat módosítja:',
        change2: 'Webes bejelentkezési jelszó',
        change3: 'Rendszer root jelszava (SSH bejelentkezési jelszó)',
        change4: 'A jelszavak visszaállításához tartsa lenyomva a BOOT gombot a IronKVM-en.',
        resetDocs: 'A részletes lépéseket a hardver dokumentációjában találod:',
        hardwareDocs: 'Sipeed NanoKVM wiki'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Wi-Fi beállítása a IronKVM-hez',
      success: 'Please check the network status of IronKVM and visit the new IP address.',
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
      },
      ssidRequired: 'Adja meg a hálózat nevét, legfeljebb 32 karakter',
      passwordLength: 'A jelszó 8–63 karakter. Nyílt hálózatnál hagyja üresen.',
      passwordOptional: 'Jelszó (nyílt hálózatnál üres)',
      lost: 'A panel nem válaszol. Lehet, hogy csatlakozott a hálózathoz, és bezárta a beállító hotspotját. Ha a hotspot újra megjelenik, a csatlakozás sikertelen volt: csatlakozzon hozzá újra, és próbálja újra.',
      done: 'A beállítás kész. Csatlakoztassa vissza ezt az eszközt a szokásos hálózathoz, és nyissa meg a panelt az új címén.'
    },
    screen: {
      viewOnly: 'Csak nézet',
      viewOnlyTip:
        'Ez a lap nem küld több billentyűzet- és egérbemenetet a gazdagépnek. A szkripteket, az egérmozgatót és a többi nézőt ez nem érinti.',
      viewOnlyOff: 'Csak nézet kikapcsolása',
      viewOnlyBlocked: 'A csak nézet be van kapcsolva, semmi sem ment a gazdagépre',
      pauseHidden: 'Szünet, ha a lap rejtett',
      pauseHiddenTip:
        'Néhány másodperccel a lap elrejtése után leállítja a képet és a hangot, és visszatéréskor újraindítja.',
      screenshot: 'Képernyőkép',
      screenshotTip: 'A gazdagép képernyőjét PNG-fájlba menti teljes rögzítési méretben.',
      screenshotFailed: 'A képernyőkép nem sikerült',
      stream: {
        ok: 'kép rendben',
        noSignal: 'nincs jel',
        failed: 'stream hiba'
      },
      codecNoWebrtcHevc: 'Ez a böngésző nem tud H.265-öt fogadni WebRTC-n keresztül',
      codecNoHevc: 'Ez a böngésző nem tudja dekódolni a H.265-öt',
      codecNote:
        'Az eszköznek egy kódolója van, így ez minden néző számára megváltoztatja a streamet. Futó WebRTC-munkamenetnél csatlakozzon újra.',
      codec: 'Kodek',
      updateFailed: 'A beállítás nem lépett érvénybe',
      scale: 'Skála',
      title: 'Képernyő',
      video: 'Videó mód',
      videoDirectTips:
        'Engedélyezze az HTTPS elemet a "Beállítások > Eszköz" menüpontban ennek a módnak a használatához',
      resolution: 'Felbontás',
      ocr: {
        title: 'Szöveg beolvasása (OCR)',
        tips: 'A szövegfelismerés ebben a böngészőben fut. Másolás előtt javíthatja a szöveget.',
        hint: 'Húzza az egeret a beolvasandó szöveg fölé. A megszakításhoz nyomja meg az Esc billentyűt.',
        noPicture: 'Várja meg a videót, majd húzza az egeret a beolvasandó szöveg fölé.',
        cancel: 'Mégse',
        language: 'Nyelv',
        languages: {
          eng: 'Angol'
        },
        preview: 'Kijelölt terület',
        capturing: 'A képernyő rögzítése...',
        loading: 'A szövegfelismerés betöltése...',
        recognizing: 'A szöveg beolvasása...',
        noText: 'A kijelölt területen nem található szöveg.',
        copy: 'Másolás',
        copied: 'Vágólapra másolva',
        copyFailed: 'Nem sikerült a vágólapra másolni',
        selectAgain: 'Újbóli kijelölés',
        unsupported:
          'Ez a böngésző nem tud szövegfelismerést futtatni. Ehhez WebAssembly SIMD szükséges, amelyet a jelenlegi böngészők támogatnak.',
        captureFailed: 'Nem sikerült rögzíteni a képernyőt.',
        outside: 'A kijelölt terület a képen kívül esik.',
        recognizeFailed: 'A szövegfelismerés sikertelen.'
      },
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
      qualityLossless: 'Legjobb',
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
      },
      directConnectionFailed: 'A videófolyam kapcsolata sikertelen'
    },
    keyboard: {
      close: 'Bezárás',
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
        sendFailed: 'Nem lett elküldve: a bemeneti kapcsolat megszakadt',
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
        saveFailed: 'Nem sikerült menteni a vezérbillentyűt',
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
      jiggler: 'Egérmozgató',
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
      touch: 'Érintőképernyős mód',
      touchShort: 'Érintés',
      absoluteStalled: 'A célgép figyelmen kívül hagyja az abszolút egeret',
      absoluteStalledDesc:
        'A célgép már nem fogadja az abszolút egér jelentéseit, így a mutató mozgásai elvesznek. A billentyűzetet ez nem érinti. Az USB helyreállítása gyakran megoldja; a relatív mód másik végpontot használ.',
      useRelative: 'Váltás relatív módra',
      direction: 'Görgő iránya',
      scrollUp: 'Mint ezen a számítógépen',
      scrollDown: 'Fordított (természetes görgetés)',
      speed: 'Görgő sebessége',
      fast: 'Gyors',
      slow: 'Lassú',
      requestPointer:
        'Relatív mód használata. Kattintson az asztalra, hogy megjelenjen az egérmutató.',
      resetHid: 'HID alaphelyzetbe állítása',
      hidOnly: {
        switchFailed: 'Nem sikerült módot váltani. Ellenőrizze a kapcsolatot, és próbálja újra.',
        title: 'Csak HID mód',
        desc: 'Ha az egér és a billentyűzet nem válaszol, és az HID alaphelyzetbe állítása nem segít, akkor az IronKVM és az eszköz közötti kompatibilitási probléma lehet. Próbálja engedélyezni az HID-Csak módot a jobb kompatibilitás érdekében.',
        tip1: 'Az HID-Csak mód engedélyezése leválasztja a virtuális U-lemezt és a virtuális hálózatot',
        tip2: 'HID-Csak módban a képrögzítés le van tiltva',
        rebuild: 'A módváltás újraépíti az USB-kapcsolatot. A IronKVM nem indul újra',
        enable: 'Engedélyezze a HID-Csak módot',
        disable: 'A HID-Csak mód letiltása'
      },
      resetHidDone: 'Az USB HID újraindítva',
      resetHidFailed: 'Az USB HID újraindítása sikertelen'
    },
    image: {
      driveLoaded: 'lemezkép betöltve',
      driveWarning: 'nézze meg a figyelmeztetéseket',
      warning: {
        missing: 'A lemezképfájlt törölték. A gazdagép a régi másolatot olvassa, amíg ki nem adja.',
        writable: 'Írható: a gazdagép módosíthatja ezt a lemezképet.',
        tooBigForCd: 'Túl nagy a CD-meghajtóhoz ({{size}}, korlát {{max}}). Használja a lemezt.',
        tooSmallForCd: 'Túl kicsi a CD-meghajtóhoz ({{size}}). Használja a lemezt.',
        empty: 'A fájl üres, valószínűleg sikertelen feltöltés vagy letöltés miatt.'
      },
      delete: 'Törlés',
      inUse: 'Használatban van. Törlés előtt adja ki.',
      retry: 'Újra',
      loadFailed: 'Nem sikerült betölteni a képek listáját',
      readOnlyLocked: 'A módosításhoz adja ki a lemezt. Kép behelyezésekor lép érvénybe.',
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
      deleteFailed: 'A törlés sikertelen',
      ventoy: {
        statusNoKernel: 'Ez a firmware nem támogatja',
        statusNotInstalled: 'Nincs telepítve',
        statusReady: 'Kész',
        statusSelected: 'Kiválasztott képfájlok: {{count}}',
        statusInDrive: 'A lemezmeghajtóban, {{size}}',
        noKernel:
          'Ennek a firmware-nek a kernele nem támogatja a device-mappert, ezért a Ventoy csak egy ilyen támogatással rendelkező kép telepítése után használható.',
        installDesc: 'A gazdagép indítása több képfájlból egyetlen lemezen, másolás nélkül.',
        install: 'Telepítés',
        installing: 'A Ventoy letöltése, kb. 20 MB. Ez néhány percig tarthat.',
        needsData: 'A Ventoyhoz olyan IronKVM kép kell, amelyen a /data partíció csatolva van.',
        uninstall: 'Eltávolítás',
        uninstallConfirm: 'Eltávolítja a Ventoy fájljait?',
        noImages: 'Nincs képfájl a Ventoy lemezhez.',
        onDisk: 'A Ventoy lemezen',
        missing: 'Hiányzik: {{file}}',
        remove: 'Eltávolítás a Ventoy lemezről',
        setHint: 'A képfájlok köre csak akkor módosítható, ha a Ventoy lemez nincs meghajtóban.',
        useAsDisk: 'Használat virtuális lemezként',
        failed: 'A Ventoy kérés sikertelen',
        secureBoot:
          'Bekapcsolt Secure Boot esetén a gazdagépnek egyszer regisztrálnia kell a Ventoy kulcsát a MokManagerben. Az ENROLL_THIS_KEY_IN_MOKMANAGER.cer kulcsfájl a VTOYEFI partíción található.',
        readOnly:
          'A gazdagép csak olvashatóként látja a lemezt, ezért a Ventoy perzisztencia és a meghajtón lévő ventoy.json nem működik.'
      },
      tips: {
        title: 'Hogyan tölts fel képeket',
        usb1: 'Csatlakoztassa a IronKVM-t a számítógépéhez USB-n keresztül.',
        usb2: 'Győződjön meg róla, hogy a virtuális lemez csatlakoztatva van (Beállítások - Virtuális lemez).',
        usb3: 'Nyissa meg a virtuális lemezt a számítógépén, és másolja a kép fájlt a virtuális lemez gyökérkönyvtárába.',
        scp1: 'Győződjön meg róla, hogy a IronKVM és a számítógépe ugyanazon a helyi hálózaton van.',
        scp2: 'Nyisson meg egy terminált a számítógépén, és használja az SCP parancsot a kép fájl feltöltésére a /data könyvtárba a IronKVM-en.',
        scp3: 'Példa: scp your-image-path root@your-ironkvm-ip:/data',
        tfCard: 'TF Kártya',
        tf1: 'Ez a módszer támogatott Linux rendszeren',
        tf2: 'Vegye ki a TF kártyát a IronKVM-ből (a TELJES verzióhoz, először szedje szét a házat).',
        tf3: 'Helyezze a TF kártyát egy kártyaolvasóba, és csatlakoztassa a számítógépéhez.',
        tf4: 'Másolja a képfájlt a TF kártya /data könyvtárába.',
        tf5: 'Helyezze vissza a TF kártyát a IronKVM-be.'
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
      close: 'Bezárás',
      empty: 'Még nincs szkript. Töltsön fel egy .sh vagy .py fájlt a panelen való futtatáshoz.',
      loadFailed: 'Nem sikerült betölteni a szkripteket',
      uploaded: 'Szkript feltöltve',
      uploadFailed: 'Nem sikerült feltölteni a szkriptet',
      started: 'A szkript elindult a háttérben',
      deleteFailed: 'Nem sikerült törölni a szkriptet',
      waitLimit: 'Várakozás a szkript befejezésére, legfeljebb {{minutes}} percig.',
      timedOut:
        'A szkript {{minutes}} percnél tovább futott, és az oldal abbahagyta a várakozást. Lehet, hogy még fut a panelen.'
    },
    terminal: {
      invalidBaud: 'Ez az átviteli sebesség nem támogatott.',
      invalidPort: 'Adjon meg egy /dev alatti eszközútvonalat, például /dev/ttyS1.',
      invalidSettings: 'Érvénytelen soros port beállítások. Ez az eszköz saját shellje.',
      disconnected: 'A kapcsolat megszakadt. Az újracsatlakozáshoz nyomja meg az Entert.',
      title: 'Terminál',
      nanokvm: 'IronKVM Terminál',
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
      no: 'Nem',
      yes: 'Igen',
      deleteConfirm: 'Törli ezt a mentett címet?',
      delete: 'Törlés',
      wake: 'Ébresztés',
      rename: 'Átnevezés',
      showMac: 'MAC-cím megjelenítése',
      showName: 'Név megjelenítése',
      requestFailed: 'Nem sikerült elérni az eszközt a parancs küldéséhez',
      deleteFailed: 'A törlés sikertelen',
      renameFailed: 'Az átnevezés sikertelen',
      title: 'Wake-on-LAN',
      sending: 'Parancs küldése...',
      sent: 'Parancs elküldve',
      input: 'Adja meg a MAC címet',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'A feltöltés sikertelen',
      uploadSuccess: 'Feltöltés kész',
      uploading: 'Feltöltés: {{file}}',
      downloadingPercent: 'Letöltés ({{percent}}): {{file}}',
      downloading: 'Letöltés: {{file}}',
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
      bootMenuPresent: 'A(z) {{file}} már az eszközön van, helyes ellenőrzőösszeggel',
      bootMenuDesc: 'A netboot.xyz ISO letöltése ellenőrzött ellenőrzőösszeggel a virtuális CD-hez'
    },
    alerts: {
      title: 'Figyelmet igényel',
      temperature: {
        warning: 'A panel {{celsius}} °C-os. Ellenőrizze, hogy kap-e levegőt.',
        critical:
          'A panel {{celsius}} °C-os, ez túl meleg. Biztosítson szellőzést vagy kapcsolja ki.'
      },
      storage: {
        warning:
          'Csak {{available}} szabad a(z) {{total}} méretből itt: {{path}}. Nagy lemezképek nem biztos, hogy elférnek.',
        critical:
          'Csak {{available}} szabad itt: {{path}}. A feltöltések, letöltések és bővítménytelepítések sikertelenek lesznek. Törölje a felesleges lemezképeket.'
      },
      vpn: 'A(z) {{name}} indításkor indulna, de nem fut, így a távoli elérés rajta keresztül nem működik.',
      openVpn: 'VPN-beállítások megnyitása',
      stream:
        'A videostream leállt. Próbáljon másik videómódot a Képernyő menüben, vagy töltse újra az oldalt.'
    },
    power: {
      resetDesc: 'Azonnal újraindítja a gazdagépet. A mentetlen munka elvész.',
      powerShortDesc:
        'Bekapcsolja a gazdagépet, vagy leállásra kéri az operációs rendszerét (ACPI).',
      powerLongDesc: 'Leállítás nélkül kényszeríti ki a gazdagép kikapcsolását.',
      hddLed: 'Lemez LED',
      hddActive: 'Aktív',
      hddIdle: 'Tétlen',
      title: 'Bekapcsolás',
      showConfirm: 'Megerősítés',
      showConfirmTip:
        'Rövid bekapcsológomb-nyomás előtt kérdezzen. Az újraindítás és a hosszú nyomás mindig kérdez.',
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
      ledConnectedFailed: 'Nem sikerült menteni a bekapcsolásjelző LED beállítását',
      powerLongConfirm:
        'Nyomva tartja a bekapcsológombot {{seconds}} mp-ig? Ez leállítás nélkül kapcsolja ki a tápot.',
      done: 'Gomb megnyomva',
      failed: 'A gombnyomás sikertelen'
    },
    settings: {
      title: 'Beállítások',
      nav: {
        system: 'Rendszer',
        network: 'Hálózat',
        access: 'Hozzáférés',
        integrations: 'Integrációk',
        boot: 'Rendszerindítás és média',
        browser: 'Ez a böngésző',
        search: 'Beállítás keresése',
        noMatch: 'Nincs egyező beállítás',
        locked:
          'Egy művelet folyamatban van. A többi oldal és a bezárás a befejezéséig nem érhető el.',
        vpnProvider: 'VPN-szolgáltató'
      },
      mcp: {
        keyNote:
          'Az MCP saját API-kulcsot használ, lent látható. Az API-kulcsok oldal kulcsai itt nem működnek.',
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
        cancelBtn: 'Mégse',
        showKey: 'Kulcs megjelenítése',
        hideKey: 'Kulcs elrejtése',
        regenerateKey: 'Új kulcs generálása'
      },
      redfish: {
        example: 'Példa',
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
      ipmi: {
        copyBeforeSave: 'Másolja ki a jelszót most. Mentés után már nem jeleníthető meg.',
        noLogin:
          'Az IPMI be van kapcsolva, de egyetlen aktív fióknak sincs IPMI-jelszava, így senki sem tud bejelentkezni. Állítson be egyet lent.',
        title: 'IPMI',
        warning:
          'Az IPMI-hitelesítés a felépítéséből adódóan gyenge. Aki eléri a kártyát és ismer egy felhasználónevet, megszerezheti a felhasználó IPMI-jelszavának hash-ét, és offline megpróbálhatja feltörni. Használjon generált jelszavakat, csak megbízható hálózaton kapcsolja be az IPMI-t, és ahol az eszköz támogatja, inkább a HTTPS feletti Redfisht használja.',
        service: 'IPMI LAN-on',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) az UDP 623-as porton a gazdagép tápellátásához és állapotához. Az IPMI 1.5 és a 0-s cipher suite el van utasítva. Kikapcsolása minden IPMI-munkamenetet lezár.',
        example: 'Példa',
        copyFailed: 'A másolás nem sikerült. Másolja kézzel.',
        ledOn: 'Elérhető az állapot, on, off, soft, cycle és reset.',
        ledOff:
          'A "Bekapcsolásjelző LED csatlakoztatva" ki van kapcsolva a tápellátás menüben, ezért a tápállapot ismeretlen. Csak a "power reset" működik: a status, on, off, soft és cycle el van utasítva.',
        accounts: 'Fiókok',
        accountsDesc:
          'Az IPMI a KVM-fiókokkal jelentkezik be, mindegyik saját, a webes jelszótól különböző IPMI-jelszóval. A rendszergazdák ADMINISTRATOR szintet kapnak. A felhasználók USER szintet: a "-L USER" kapcsolóval olvashatják a tápállapotot, de nem módosíthatják.',
        passwordSet: 'IPMI-jelszó beállítva',
        passwordNotSet: 'Nincs IPMI-jelszó: IPMI-n nem tud bejelentkezni',
        nameTooLong: 'A név hosszabb 16 karakternél, amit az IPMI nem enged',
        accountDisabled: 'A fiók le van tiltva',
        setPassword: 'Jelszó beállítása',
        changePassword: 'Jelszó módosítása',
        remove: 'Eltávolítás',
        removeConfirmTitle: 'Eltávolítja {{user}} IPMI-jelszavát?',
        removeConfirmDesc:
          'A fiók többé nem tud IPMI-n bejelentkezni, és IPMI-munkamenetei véget érnek.',
        passwordTitle: '{{user}} IPMI-jelszava',
        passwordDesc:
          '12 és 20 közötti számú nyomtatható ASCII-karakter, eltérő a webes jelszótól. Az IPMI megköveteli, hogy a kártya visszaolvasható formában tárolja a jelszót, ezért olyat használjon, amelyet sehol máshol nem használ. Mentés előtt másolja ki: többé nem jelenik meg.',
        passwordPlaceholder: 'IPMI-jelszó',
        generate: 'Generálás',
        copy: 'Másolás',
        save: 'Mentés',
        passwordLength: 'Használjon 12 és 20 közötti számú karaktert.',
        passwordChars: 'Csak nyomtatható ASCII-karaktereket használjon.',
        saved: 'IPMI-jelszó mentve',
        failed: 'Az IPMI-művelet nem sikerült',
        okBtn: 'Megerősítés',
        cancelBtn: 'Mégse'
      },
      ssh: {
        service: 'SSH-kiszolgáló',
        serviceDesc: 'Az sshd indítása most és minden rendszerindításkor',
        failed: 'Az SSH-beállítások nem tölthetők be',
        rootDefault: 'A root még mindig a gyári jelszót használja',
        rootEmpty: 'A rootnak nincs jelszava',
        rootWarning:
          'Aki eléri a konzolt vagy az SSH-t, root-ként beléphet. Állítson be jelszót itt: {{account}} > {{password}}. Az eszköz tulajdonosánál ez a root jelszavát is beállítja.',
        connection: 'Kapcsolat',
        command: 'Belépés root-ként',
        port: 'Port',
        viaVpn: '{{name}} útján',
        notRunning: 'Az sshd nem fut. A csatlakozáshoz kapcsolja be az SSH-kiszolgálót.',
        hostKeys: 'Gépkulcs-ujjlenyomatok',
        hostKeysDesc: 'Vesse össze őket azzal, amit az ssh az első kapcsolódáskor mutat.',
        noHostKeys: 'Még nincsenek gépkulcsok. Az sshd az első indításkor hozza létre őket.',
        keys: 'Engedélyezett kulcsok',
        keysDesc:
          'Nyilvános kulcsok, amelyekkel root-ként be lehet lépni. Az adatpartíción tárolódnak, így a frissítések megtartják őket.',
        noKeys: 'Még nincsenek engedélyezett kulcsok.',
        noComment: 'nincs megjegyzés',
        addPlaceholder:
          'Illesszen be egy nyilvános kulcsot, például a ~/.ssh/id_ed25519.pub tartalmát',
        add: 'Kulcs hozzáadása',
        added: 'Kulcs hozzáadva',
        removed: 'Kulcs eltávolítva',
        deleteConfirm: 'Eltávolítja ezt a kulcsot?',
        deleteConfirmDesc: 'Többé nem tud belépni. A nyitott munkamenetek nyitva maradnak.',
        invalidKey: 'Ez nem nyilvános kulcs. Illesszen be egyetlen sort egy .pub fájlból.',
        keyOptions:
          'Az olyan beállításokat tartalmazó kulcsok, mint a command= vagy a from=, itt nem fogadhatók el.',
        duplicateKey: 'Ez a kulcs már engedélyezett.',
        lastKey:
          'Az utolsó kulcs nem távolítható el, amíg a csak kulcsos belépés be van kapcsolva.',
        keysOnly: 'Csak kulcsok',
        keysOnlyDesc:
          'A jelszavas és a keyboard-interactive belépés kikapcsolása. A nyitott munkamenetek nyitva maradnak.',
        keysOnlyNeedsKey:
          'Előbb adjon hozzá egy engedélyezett kulcsot, különben senki sem tudna belépni.',
        keysOnlyOn: 'Jelszavas belépés kikapcsolva',
        keysOnlyOff: 'Jelszavas belépés bekapcsolva',
        notHonoured:
          'Ennek a képfájlnak az sshd-je nem olvassa ezt a beállítást, így a jelszavas belépés bekapcsolva marad.',
        reloadFailed:
          'Mentve, de az sshd nem tölthető újra. Az sshd következő indításakor lép érvénybe.',
        notApplied:
          'Az sshd még elfogad jelszavakat. A beállítás alkalmazásához kapcsolja ki, majd be az SSH-kiszolgálót.',
        changePort: 'Módosítás',
        portConfirm: 'Módosítja az SSH-portot erre: {{port}}?',
        portConfirmDesc:
          'A jelenlegi SSH-munkamenetek nyitva maradnak. Az új kapcsolatoknak a(z) {{port}} portot kell használniuk. Győződjön meg róla, hogy a tűzfal engedi.',
        portChanged: 'Az SSH-port új értéke: {{port}}',
        portInvalid: 'Adjon meg egy portot 1 és 65535 között.',
        portReserved: 'Ezt a portot maga az IronKVM használja. Válasszon másikat.',
        portInUse: 'Ezen a porton már egy másik program figyel az IronKVM-en.',
        portNotHonoured:
          'Ennek a képfájlnak az sshd-je nem olvassa ezt a beállítást, így a port nem változik.'
      },
      vnc: {
        address: 'Cím',
        certHint:
          'A VeNCrypt X509Plain az eszköz önaláírt tanúsítványát használja, ezért a kliens az első csatlakozáskor figyelmeztet. Fogadja el, vagy mentse a tanúsítványt az oldal HTTPS-címéről, és adja át a TigerVNC-nek a -X509CA=<fájl> kapcsolóval.',
        title: 'VNC',
        service: 'VNC-kiszolgáló',
        serviceDesc:
          'Egy VNC-kliens, például a TigerVNC vagy a Remmina, láthatja és vezérelheti a gazdagépet. A kliensnek támogatnia kell a Tight kódolást. Egyszerre egy munkamenet.',
        credentials:
          'Jelentkezzen be egy KVM-fiókkal. A kapcsolatot a kártya TLS-tanúsítványa titkosítja (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'A TCP-port, amelyen a kiszolgáló figyel.',
        maxFps: 'Képkocka-korlát',
        maxFpsDesc: 'A kliensnek küldött képkockák legnagyobb száma másodpercenként.',
        vncAuth: 'Egyszerű VNC-hitelesítés',
        vncAuthDesc: 'VeNCrypt nélküli klienseknek. Fiók helyett egy külön VNC-jelszót ellenőriz.',
        vncAuthWarning:
          'Az egyszerű VNC-hitelesítés nem titkosítja a kapcsolatot. A hálózati útvonalon bárki láthatja a képernyőt és a billentyűleütéseket. Csak megbízható hálózaton használja.',
        password: 'VNC-jelszó',
        passwordSet: 'Van beállított jelszó. A módosításhoz írjon be egy újat.',
        passwordInvalid: 'A VNC-jelszó hossza 6 és 8 karakter között legyen.',
        save: 'Mentés',
        saved: 'Beállítások mentve',
        state: 'Állapot',
        listening: 'Figyel a(z) {{port}} porton',
        notListening: 'Nem figyel',
        noSession: 'Nincs nyitott munkamenet',
        client: 'Kliens',
        user: 'Felhasználó',
        method: 'Hitelesítés',
        methodVencrypt: 'Fiók TLS-en keresztül',
        methodVnc: 'VNC-jelszó',
        since: 'Csatlakozva ekkor óta',
        resolution: 'Felbontás',
        framesSent: 'Elküldött képkockák',
        lastError: 'Az utolsó munkamenet véget ért: {{error}}',
        refresh: 'Frissítés',
        disconnect: 'Bontás',
        disconnectConfirmTitle: 'Befejezi a VNC-munkamenetet?',
        disconnectConfirmDesc:
          'A kliens kapcsolata azonnal megszakad, és minden lenyomva tartott billentyű és gomb felengedésre kerül.',
        failed: 'A VNC-művelet sikertelen',
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
        actionReset: 'Újraindítás',
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
        failed: 'A watchdog művelete nem sikerült',
        powerNeedsLed:
          'A tápciklushoz be kell kapcsolni a „Táp LED csatlakoztatva” beállítást a tápmenüben.',
        noLedConfirmTitle: 'Bekapcsolja a watchdogot táp LED nélkül?',
        noLedConfirmDesc:
          'A panel nem látja, mikor van kikapcsolva a gazdagép, ezért mindig bekapcsoltnak tekinti. Ha leállítja a gazdagépet, a watchdog az időkorlát után megnyomja az újraindítást. Ennek elkerüléséhez csatlakoztassa a táp LED-et.',
        noLedConfirmOk: 'Bekapcsolás',
        cancel: 'Mégse'
      },
      media: {
        title: 'Virtuális média',
        description:
          'Az eszköztár Média párbeszédablakának beállítása. A képfájlok csatolása, hozzáadása és a Ventoy-készlet kiválasztása a párbeszédablakban történik.',
        ejectFirst:
          'A Ventoy-lemez egy meghajtóban van. Az eltávolításhoz adja ki a Média párbeszédablakban.'
      },
      netboot: {
        title: 'Hálózati rendszerindítás',
        isoDownload: 'Letöltés',
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
        title: 'IronKVM Névjegy',
        information: 'Információ',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Alkalmazás verzió',
        applicationTip: 'IronKVM webalkalmazás verziója',
        image: 'Képfájl verzió',
        imageTip: 'IronKVM kártyakép és a NanoKVM rendszerkép, amelyre épül',
        kernel: 'Kernelverzió',
        kernelTip: 'A jelenleg futó Linux-kernel kiadása',
        deviceKey: 'Eszköz kulcs',
        videoMemory: 'Videomemória',
        videoMemoryTip:
          'Videorögzítésre fenntartott memória. A rendszer többi része nem használja.',
        videoMemoryGenerations_one: '{{count}} korábbi IronKVM-munkamenet foglal videomemóriát',
        videoMemoryGenerations_other: '{{count}} korábbi IronKVM-munkamenet foglal videomemóriát',
        videoMemoryReboot: 'A felszabadításhoz indítsa újra.',
        community: 'Közösség',
        hostname: 'Gazdanév',
        hostnameUpdated: 'Gazdanév frissítve. Az alkalmazáshoz indítsa újra.',
        ipType: {
          Wired: 'Vezetékes',
          Wireless: 'Vezeték nélküli',
          Other: 'Egyéb'
        },
        hostnameInvalid:
          'Betűket, számjegyeket és kötőjeleket használjon, pontokkal elválasztott részenként legfeljebb 63-at. Rész elején vagy végén nem lehet kötőjel.',
        hostnameFailed: 'Nem sikerült módosítani a gépnevet',
        editHostname: 'Gépnév szerkesztése',
        docs: 'Dokumentáció',
        hardware: 'Hardver',
        hardwareFaq: 'Hardver GYIK',
        disclaimer:
          'IronKVM: megerősített közösségi firmware a Sipeed NanoKVM-hez. Nem áll kapcsolatban a Sipeeddel.',
        basedOn: 'NanoKVM {{version}} alapján'
      },
      preferences: {
        title: 'Preferenciák'
      },
      performance: {
        title: 'Teljesítmény'
      },
      appearance: {
        thisBrowser: 'Ez a böngésző',
        thisBrowserDesc:
          'Csak ebben a böngészőben tárolva. A többi böngésző saját beállítást használ.',
        deviceWide: 'Eszköz',
        deviceWideDesc: 'Az eszközön tárolva. Mindenkire vonatkozik, aki megnyitja.',
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
          15: '15 mp',
          30: '30 mp',
          60: '1 perc',
          180: '3 perc',
          300: '5 perc',
          600: '10 perc',
          1800: '30 perc',
          3600: '1 óra'
        },
        sections: {
          video: 'Videó',
          usb: 'USB',
          frontPanel: 'Előlap'
        },
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
          tip: 'Ennek a funkciónak az engedélyezése lerövidítheti az SD-kártya élettartamát!',
          active: 'Aktív - {{used}} / {{total}}',
          inactive: 'Beállítva, de nincs használatban'
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
        hidOnly: 'HID-Csak mód',
        hidOnlyDesc:
          'A virtuális eszközök emulálásának leállítása, csak az alapvető HID vezérlés megtartásával',
        disk: 'Virtuális lemez',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Virtuális hálózat',
        networkDesc: 'Virtuális hálózati kártya csatlakoztatása a távoli gazdagépen',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Gazdagép:',
          description:
            'Privát hálózati kapcsolat a távoli gazdagéppel az USB-kábelen keresztül. A gazdagép átjáró és DNS nélküli címet kap, így a IronKVM-en keresztül nem éri el a helyi hálózatot.',
          mode: 'Protokoll',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (NCM nélküli gazdagépekhez)',
          rndis: 'RNDIS (már nem választható)',
          rndisNote:
            'Ez a kapcsolat RNDIS-t használ, amely már nem választható. Válassza az NCM-et vagy az ECM-et.',
          subnet: 'Alhálózat',
          subnetDesc:
            'Privát IPv4-hálózat, /24 és /30 között. A IronKVM az első címet kapja, a gazdagép a másodikat.',
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
          'USB soros portot jelenít meg a távoli gazdagépen, amelyen át bejelentkezhet erre a IronKVM-re, ha a hálózat nem érhető el',
        consoleTip:
          'Bárki, aki a távoli gazdagépet vezérli, bejelentkezési promptot kap ehhez a IronKVM-hez. Az engedélyezés előtt állítson be erős jelszót (Fiók - Jelszó módosítása).',
        usbApply: {
          changed: 'Módosítva',
          discard: 'Elvetés',
          pending: 'A módosítások még nincsenek alkalmazva.'
        },
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
        rebootDesc: 'Biztos, hogy újra akarja indítani a IronKVM-t?',
        okBtn: 'Igen',
        cancelBtn: 'Nem',
        rebootFailed: 'Az újraindítás sikertelen'
      },
      network: {
        title: 'Hálózat',
        wifi: {
          disconnectBtn: 'Leválasztás',
          disconnectWarning:
            'Ha ezen a Wi-Fi hálózaton éri el a IronKVM-et, ez az oldal elveszíti a kapcsolatot.',
          disconnected: 'Wi-Fi leválasztva',
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
          waitingHttp: 'Visszaváltás http-re. Ha az oldal nem nyílik meg magától, töltse újra.',
          failed: 'Nem sikerült módosítani a HTTPS beállítást',
          enableConfirm: 'Bekapcsolja a HTTPS-t?',
          disableConfirm: 'Kikapcsolja a HTTPS-t?',
          confirmDesc:
            'Ez kijelentkeztet, és újraindítja az eszköz szerverét, ami körülbelül két percig tart. Ezután az oldal megnyitja: {{url}}.',
          confirmOk: 'Folytatás',
          confirmCancel: 'Mégse'
        },
        ethernet: {
          title: 'IP-cím',
          description: 'Állítsa be, hogyan kapja a IronKVM a címét a vezetékes hálózaton',
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
          applyTitle: 'Megváltoztatja a IronKVM címét?',
          applyWarning:
            'A kapcsolat ezzel az oldallal megszakad. A IronKVM alkalmazza az új címet, és {{seconds}} másodpercet vár arra, hogy elérje azon a címen. Az elérés megtartja a módosítást. Ha semmi sem éri el, a IronKVM visszaállítja a korábbi beállításokat.',
          applyConfirm: 'Alkalmaz',
          applyCancel: 'Mégse',
          applyFailed: 'A cím alkalmazása nem sikerült',
          trialTitle: 'Várakozás a megerősítésre',
          trialDhcp: 'A IronKVM címet kér a DHCP-től.',
          trialStatic: 'A IronKVM most a következő címen érhető el: {{address}}.',
          trialInstruction:
            'Nyissa meg a IronKVM-et az új címén, és jelentkezzen be, ha kéri. Az elérés megtartja a módosítást. Ha {{seconds}} másodpercen belül semmi sem éri el a IronKVM-et, visszaállítja a korábbi beállításokat.',
          trialOpen: 'Az új cím megnyitása',
          trialKeep: 'Beállítások megtartása',
          trialKept: 'Az új cím mentve',
          trialKeepFailed: 'A beállításokat nem sikerült megtartani',
          trialGone: 'A módosítás már vissza lett vonva. Próbálja újra.',
          unsaved: 'Nem mentett módosítások'
        },
        dns: {
          title: 'DNS',
          description: 'DNS-kiszolgálók beállítása a IronKVM számára',
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
        connect: 'Csatlakozás',
        connectDesc:
          'Csatlakozás a(z) {{name}} hálózathoz. Kikapcsolva a szolgáltatás leállítása nélkül bontja a kapcsolatot.',
        kvmUrl: 'KVM címe',
        moreTip: 'További műveletek',
        restartTip: 'Újraindítás',
        stopTip: 'Leállítás',
        updateTip: 'Frissítés erre: {{version}}',
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
          tip: 'Ha a démonnak kevés a memóriája, próbálja bekapcsolni a swapot. Ezt a "Beállítások > Teljesítmény" alatt lehet megadni.'
        },
        copy: 'Másolás',
        copied: 'Hivatkozás másolva',
        copyFailed: 'Nem sikerült másolni a hivatkozást. Jelölje ki, és másolja kézzel.',
        open: 'Megnyitás',
        checkAgain: 'Ellenőrzés újra',
        notSignedIn:
          'Még nincs bejelentkezve. Fejezze be a bejelentkezést a hivatkozáson, majd ellenőrizze újra.',
        checkFailed: 'Nem sikerült ellenőrizni a bejelentkezés állapotát',
        loginWaiting:
          'Az oldal néhány másodpercenként ellenőrzi, és a bejelentkezés után folytatja.',
        uninstallFailed: 'Az eltávolítás sikertelen',
        loginFailed: 'A bejelentkezés sikertelen'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Letöltés a',
        package: 'telepítési csomag',
        unzip: 'és kicsomagolás',
        notLogin:
          'Az eszköz még nincs kötve. Kérem, jelentkezzen be és kösse az eszközt a fiókjához.',
        urlPeriod: 'Ez az url 10 percig érvényes',
        login: 'Bejelentkezés',
        logout: 'Kijelentkezés',
        logoutDesc: 'Biztos, hogy ki szeretne jelentkezni?',
        manualIntro: 'Vagy telepítsd kézzel SSH-n keresztül:',
        copyBinaries: 'Másold a tailscale és tailscaled fájlt az IronKVM {{dir}} könyvtárába',
        linksFile: 'Ugyanebben a könyvtárban hozz létre egy links nevű fájlt ezzel a két sorral:',
        rebootRefresh: 'Indítsd újra az IronKVM-et, majd frissítsd ezt az oldalt'
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
        logout: 'Regisztráció törlése',
        logoutDesc:
          'A regisztráció törlése eltávolítja ezt a társat a NetBird-fiókjából, és törli itt a konfigurációját. Az újbóli csatlakozáshoz beállítókulcs vagy SSO-bejelentkezés kell, és a társ új IP-címet kaphat. Folytatja?',
        joinFailed: 'Nem sikerült csatlakozni a hálózathoz'
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
            'Az SHA-512 csak azt ellenőrzi, hogy a csomag megfelel-e a kiszolgáló által biztosított jegyzéknek. Nem igazolja, hogy a csomag hivatalos IronKVM-kiadás. Egy hibás vagy rosszindulatú kiszolgáló használhatatlanná teheti az eszközt, adatvesztést okozhat, vagy veszélyeztetheti a rendszert.',
          confirm: 'Használat mindenképpen',
          useSipeed: 'A hivatalos Sipeed kiszolgáló használata',
          previewDisabled:
            'Az előzetes frissítések nem érhetők el, amíg egyéni frissítési kiszolgáló van engedélyezve.'
        },
        offline: {
          chooseFile: 'Fájl kiválasztása',
          installing: 'Feltöltés kész. Telepítés...',
          noFile: 'Nincs kiválasztott fájl',
          title: 'Offline frissítések',
          desc: 'Frissítés helyi telepítőcsomaggal',
          upload: 'Feltöltés',
          checksumPlaceholder: 'SHA-256 ellenőrzőösszeg (opcionális)',
          invalidChecksum: 'A SHA-256 ellenőrzőösszegnek 64 hexadecimális karakterből kell állnia.',
          checksumMismatch: 'Az SHA-256 ellenőrzése sikertelen. Lehet, hogy a csomag sérült.',
          invalidName: 'Érvénytelen fájlnévformátum. Kérjük, töltse le a GitHub kiadásaiból.',
          updateFailed: 'Frissítés sikertelen. Kérem, próbálja újra.'
        },
        updateTo: 'Frissítés erre: {{version}}',
        updateConfirmDesc:
          'Az eszköz telepíti a frissítést és újraindítja a szerverét. Az oldal újratöltődik, amikor a szerver ismét elérhető.',
        releaseNotes: 'Kiadási megjegyzések'
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
        mcpNote:
          'Ezek a kulcsok nem működnek az MCP-hez, amelynek saját kulcsa van az MCP oldalon.',
        metricsUrl: 'Metrikák URL-je',
        monitoring: 'Megfigyelés',
        monitoringDesc:
          'A Prometheus az ezen az oldalon létrehozott API-kulccsal olvassa a metrikákat, Bearer tokenként küldve. Bármely szerepkör olvashatja őket.',
        scrapeConfig: 'Prometheus scrape konfiguráció',
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
        kvmDescription: 'Működtesse a távoli gazdagépet az IronKVM segítségével.',
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
          restoring: 'PicoClaw visszaállítása',
          ready: 'Runtime kész',
          stopped: 'Runtime leállt',
          blockedByMCP: 'A külső MCP-vezérlés aktív',
          readyBlockedByMCP: 'A runtime fut, de jelenleg külső MCP vezérli az eszköz bevitelét.',
          readyWithoutControl:
            'A runtime fut. Újracsatlakozás előtt add át az eszközvezérlést a PicoClaw-nak.',
          unavailable: 'Runtime nem érhető el',
          configError: 'Konfigurációs hiba'
        },
        transport: {
          connecting: 'Csatlakozás',
          connected: 'Csatlakoztatva',
          disconnected: 'Leválasztva',
          reconnect: 'Újracsatlakozás',
          reconnectDescription: 'Újracsatlakozás a futó PicoClaw-munkamenethez.',
          reconnectBlocked: 'A PicoClaw-nak eszközvezérlés kell az újracsatlakozáshoz.'
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
        picoclawDescription:
          'A PicoClaw billentyűzet- és egérbevitelt küldhet. A kézi bevitel szünetelhet.',
        mcp: 'Eszközvezérlés: külső MCP',
        mcpDescription: 'A külső MCP írhat az eszközre. A PicoClaw nem veszi át a bevitelt.',
        off: 'Eszközvezérlés: kikapcsolva',
        offDescription:
          'Az MI nem küld billentyűzet- vagy egérbevitelt. A kézi vezérlés továbbra is elérhető.',
        transitioning: 'Eszközvezérlés: váltás',
        transitioningDescription: 'Az eszközvezérlés szinkronizálódik. Kérlek, várj.',
        grant: 'Vezérlés átadása',
        release: 'Vezérlés feloldása',
        releasing: 'Feloldás...',
        switching: 'Váltás...',
        releasingLabel: 'Eszközvezérlés: feloldás',
        releasingDescription:
          'Az eszközvezérlés visszaadása folyamatban. A PicoClaw leállította a folyamatban lévő bevitelt.',
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
        switchFromMCP: 'Váltás PicoClaw-ra és indítás',
        takeoverAndStart: 'Átvétel és indítás'
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
    upstream: {
      check: 'Frissítések keresése',
      updateTo: 'Frissítés erre: {{version}}',
      confirm: 'Frissíted a(z) {{name}} összetevőt erre: {{version}}?',
      confirmDesc:
        'Az új kiadás a GitHubról töltődik le, és az általa közzétett ellenőrzőösszegekkel ellenőrizzük. Ha valami hibázik, a jelenlegi verzió marad.',
      ok: 'Frissítés',
      upToDate: 'Naprakész',
      builtIn: 'beépített',
      checkFailed: 'Nem sikerült frissítéseket keresni: {{error}}',
      unverifiable: 'A(z) {{version}} verzió nem érhető el: {{reason}}',
      inUse: 'Most nem frissíthető: {{reason}}',
      running: 'Frissítés erre: {{version}}...',
      done: '{{name}} frissítve erre: {{version}}',
      failed: 'Az utolsó frissítés sikertelen volt: {{error}}'
    },
    menu: {
      mediaAdd: 'Képfájl hozzáadása',
      mediaMoreOptions: 'További beállítások',
      mediaSettings: 'Média beállításai',
      collapse: 'Menü összecsukása',
      expand: 'Bontsa ki a menüt',
      more: 'Továbbiak',
      media: 'Média',
      tools: 'Eszközök',
      text: 'Szöveg',
      advanced: 'Speciális',
      mediaMounted: 'Csatolva',
      mediaLibrary: 'Könyvtár',
      textToHost: 'A gazdagép felé',
      textFromHost: 'A gazdagéptől'
    },
    ion: {
      checking: 'Videomemória ellenőrzése az adatfolyam indítása előtt...',
      warn: 'Kevés a videomemória. Egyetlen szerver-újraindítás elfogyasztaná. Indítsa újra, amikor alkalmas.',
      criticalTitle: 'Nincs elég videomemória az adatfolyam indításához',
      criticalBody:
        'A videó indítása elfogyasztaná a fenntartott memóriát, és leállítaná a szervert. Minden más funkció továbbra is működik, beleértve a tápellátás-vezérlést és az újraindítást. Ezt a memóriát csak a IronKVM újraindítása szabadítja fel.',
      criticalContinue: 'Videó indítása mégis',
      criticalReboot: 'IronKVM újraindítása',
      criticalRebooting: 'Újraindítás...'
    }
  }
};

export default hu;
