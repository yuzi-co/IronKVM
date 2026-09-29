const pl = {
  translation: {
    feedback: {
      enabled: '{{name}} włączone',
      disabled: '{{name}} wyłączone',
      failed: 'Żądanie nie powiodło się. Spróbuj ponownie.',
      network: 'Nie można połączyć się z urządzeniem. Sprawdź połączenie i spróbuj ponownie.',
      saved: 'Zapisano',
      timeout: 'Urządzenie zbyt długo nie odpowiadało. Spróbuj ponownie.'
    },
    common: {
      copy: 'Kopiuj',
      copied: 'Skopiowano',
      copyFailed: 'Nie udało się skopiować. Zaznacz tekst i skopiuj go ręcznie.',
      notUpdating: 'Brak aktualizacji: ostatnie odświeżenie nie powiodło się.',
      off: 'Wyłączone',
      running: 'Działa',
      save: 'Zapisz',
      cancel: 'Anuluj',
      delete: 'Usuń',
      remove: 'Usuń'
    },
    head: {
      desktop: 'Zdalny pulpit',
      login: 'Logowanie',
      changePassword: 'Zmień Hasło',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Hasło zmienione. Zaloguj się nowym hasłem.',
      cookieRejected:
        'Przeglądarka odmówiła zapisania sesji. Pliku cookie pozostałego po poprzedniej sesji HTTPS nie można zastąpić przez zwykłe http. Wyczyść pliki cookie dla tego adresu lub otwórz okno prywatne i zaloguj się ponownie.',
      login: 'Logowanie',
      placeholderUsername: 'Wprowadź nazwę użykownika',
      placeholderPassword: 'wprowadź hasło',
      placeholderCurrentPassword: 'Obecne hasło',
      placeholderPassword2: 'wprowadź hasło ponownie',
      noEmptyUsername: 'nazwa użykownika nie może być pusta',
      noEmptyPassword: 'hasło nie może być puste',
      passwordLength: 'Hasło musi mieć od 8 do 72 znaków',
      noAccount:
        'Nie udało się uzyskać informacji o użytkowniku, odśwież stronę lub zresetuj hasło',
      invalidUser: 'Błędne hasło lub nazwa użykownika',
      locked: 'Zbyt wiele loginów, spróbuj ponownie później',
      globalLocked: 'System chroniony, spróbuj ponownie później',
      error: 'niespodziewany błąd',
      invalidCurrentPassword: 'Obecne hasło jest nieprawidłowe',
      changePassword: 'Zmień Hasło',
      changePasswordDesc:
        'Dla bezpieczeństwa Twojego urządzenia, proszę zmień hasło do logowania w sieci.',
      differentPassword: 'hasła nie zgadzają się',
      illegalUsername: 'nazwa użytkownika zawiera niedozwolone znaki',
      illegalPassword: 'hasło zawiera niedozwolone znaki',
      forgetPassword: 'Zapomiałeś hasła?',
      ok: 'Ok',
      cancel: 'Anuluj',
      loginButtonText: 'Zaloguj się',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the IronKVM for 10 seconds.',
        reset3: 'Domyślne konto web:',
        reset4: 'Domyślne konto SSH:',
        change1: 'Pamiętaj, że ta operacja zmieni następujące hasła:',
        change2: 'Hasło logowania web',
        change3: 'Hasło roota systemu (hasło logowania SSH)',
        change4: 'Aby zresetować hasła, naciśnij i przytrzymaj przycisk BOOT na IronKVM.',
        resetDocs: 'Szczegółowe kroki znajdziesz w dokumentacji sprzętu:',
        hardwareDocs: 'Wiki Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Skonfiguruj Wi-Fi dla IronKVM',
      success: 'Proszę podejść do urządzenia, aby sprawdzić stan sieci IronKVM.',
      failed: 'Operacja nie powiodła się, spróbuj ponownie.',
      invalidMode:
        'Bieżący tryb nie obsługuje konfiguracji sieci. Przejdź do swojego urządzenia i włącz tryb konfiguracji Wi-Fi.',
      confirmBtn: 'Ok',
      finishBtn: 'Zakończono',
      ap: {
        authTitle: 'Wymagane uwierzytelnienie',
        authDescription: 'Aby kontynuować, wprowadź AP hasło',
        authFailed: 'Nieprawidłowe hasło AP',
        passPlaceholder: 'AP hasło',
        verifyBtn: 'Sprawdź'
      },
      ssidRequired: 'Wpisz nazwę sieci, maksymalnie 32 znaki',
      passwordLength: 'Hasło ma od 8 do 63 znaków. W przypadku sieci otwartej zostaw je puste.',
      passwordOptional: 'Hasło (puste dla sieci otwartej)',
      lost: 'Płytka przestała odpowiadać. Mogła połączyć się z siecią i zamknąć swój hotspot konfiguracyjny. Jeśli hotspot pojawi się ponownie, połączenie się nie udało: połącz się z nim ponownie i spróbuj jeszcze raz.',
      done: 'Konfiguracja zakończona. Połącz to urządzenie z powrotem ze swoją zwykłą siecią i otwórz płytkę pod jej nowym adresem.'
    },
    screen: {
      codecNoWebrtcHevc: 'Ta przeglądarka nie odbiera H.265 przez WebRTC',
      codecNoHevc: 'Ta przeglądarka nie dekoduje H.265',
      codecNote:
        'Płytka ma jeden koder, więc zmienia to strumień dla wszystkich widzów. Połącz się ponownie, aby zastosować to w trwającej sesji WebRTC.',
      codec: 'Kodek',
      updateFailed: 'Ustawienie nie zostało zastosowane',
      scale: 'Skala',
      title: 'Ekran',
      video: 'Tryb wideo',
      videoDirectTips: 'Włącz HTTPS w „Ustawienia > Urządzenie”, aby korzystać z tego trybu',
      resolution: 'Rozdzielczość',
      ocr: {
        title: 'Odczytaj tekst (OCR)',
        tips: 'Tekst jest rozpoznawany w tej przeglądarce. Możesz go poprawić przed skopiowaniem.',
        hint: 'Przeciągnij nad tekstem do odczytania. Naciśnij Esc, aby anulować.',
        noPicture: 'Poczekaj na obraz, a następnie przeciągnij nad tekstem do odczytania.',
        cancel: 'Anuluj',
        language: 'Język',
        languages: {
          eng: 'Angielski'
        },
        preview: 'Zaznaczony obszar',
        capturing: 'Przechwytywanie ekranu...',
        loading: 'Wczytywanie rozpoznawania tekstu...',
        recognizing: 'Odczytywanie tekstu...',
        noText: 'Nie znaleziono tekstu w zaznaczonym obszarze.',
        copy: 'Kopiuj',
        copied: 'Skopiowano do schowka',
        copyFailed: 'Nie udało się skopiować do schowka',
        selectAgain: 'Zaznacz ponownie',
        unsupported:
          'Ta przeglądarka nie może uruchomić rozpoznawania tekstu. Wymaga ono WebAssembly SIMD, które obsługują obecne przeglądarki.',
        captureFailed: 'Nie udało się przechwycić ekranu.',
        outside: 'Zaznaczony obszar leży poza obrazem.',
        recognizeFailed: 'Rozpoznawanie tekstu nie powiodło się.'
      },
      controlRegion: {
        title: 'Kalibracja myszy',
        description:
          'Użyj tego ustawienia, gdy kontrolowane urządzenie korzysta z rozdzielczości innej niż 16:9, a kursor jest przesunięty w poziomie lub w pionie.',
        off: 'Wyłączona',
        auto: 'Automatyczna',
        autoWarning:
          'Kalibracja może się nie powieść, gdy aplikacja użytkownika ma całkowicie czarne tło.',
        manual: 'Ręczna',
        selectedResolution: 'Rozdzielczość zaznaczonego obszaru',
        unused: 'Nieużywane',
        originalResolution: 'Oryginalna rozdzielczość',
        selectResolution: 'Wybierz oryginalną rozdzielczość',
        addResolution: 'Dodaj niestandardową rozdzielczość',
        add: 'Dodaj',
        duplicateResolution: 'Ta rozdzielczość już istnieje.',
        width: 'Szerokość',
        height: 'Wysokość',
        apply: 'Oblicz i zastosuj',
        invalidResolution: 'Po uruchomieniu wideo wprowadź prawidłową oryginalną rozdzielczość.',
        select: 'Zaznacz obszar',
        clear: 'Przywróć automatyczne wykrywanie',
        saveFailed: 'Nie udało się zapisać obszaru wejściowego.',
        tooSmall: 'Zaznaczony obszar jest zbyt mały.',
        previewUnavailable: 'Podgląd niedostępny',
        clearConfirm: 'Przywrócić automatyczne wykrywanie czarnych obramowań?',
        dragHint: 'Przeciągnij, aby zaznaczyć obszar pulpitu zdalnego',
        finish: 'Gotowe',
        confirm: 'Potwierdź',
        cancel: 'Anuluj'
      },
      auto: 'Automatyczny',
      autoTips:
        'W określonych rozdzielczościach może wystąpić rozrywanie ekranu lub przesunięcie myszy. Rozważ dostosowanie rozdzielczości zdalnego hosta lub wyłącz tryb automatyczny.',
      fps: 'FPS',
      customizeFps: 'Personalizuj',
      quality: 'Jakość',
      qualityLossless: 'Najlepsza',
      qualityHigh: 'Wysoki',
      qualityMedium: 'Średni',
      qualityLow: 'Niski',
      frameDetect: 'Wykrywanie klatek',
      frameDetectTip:
        'Obliczanie różnicy między klatkami. Zatrzymaj transmisję strumienia wideo, gdy na ekranie zdalnego hosta nie zostaną wykryte żadne zmiany.',
      resetHdmi: 'Resetuj HDMI',
      mixedH264: {
        title: 'Konflikt strumieni H.264',
        description:
          'Strumienie H.264 Direct i H.264 WebRTC są używane jednocześnie. Może to powodować rozrywanie obrazu lub uszkodzenie wideo. Używaj tylko jednego trybu H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Połączenie WebRTC nie powiodło się',
        description: 'Sprawdź połączenie sieciowe lub zmień tryb wideo.'
      },
      captureStatus: {
        hdmiError: 'Błąd obrazu HDMI',
        unsupportedResolution: 'Bieżąca rozdzielczość nie jest obsługiwana',
        retrieving: 'Pobieranie obrazu...',
        changingResolution: 'Przełączanie rozdzielczości...',
        updateFailed: 'Nie można teraz zaktualizować obrazu',
        videoError: 'Błąd wyświetlania wideo',
        noHdmi: 'Nie wykryto sygnału HDMI',
        unavailable: 'Nie można teraz wyświetlić obrazu'
      },
      directConnectionFailed: 'Nie udało się połączyć ze strumieniem wideo'
    },
    keyboard: {
      close: 'Zamknij',
      title: 'Klawiatura',
      paste: 'Wklej',
      tips: 'Wpisuje tekst na hoście jako naciśnięcia klawiszy. Wybierz układ klawiatury używany przez host.',
      placeholder: 'Proszę wprowadzić coś',
      submit: 'Prześlij',
      virtual: 'Klawiatura',
      readClipboard: 'Czytaj ze schowka',
      clipboardPermissionDenied:
        'Odmowa dostępu do schowka. Zezwól na dostęp do schowka w przeglądarce.',
      clipboardReadError: 'Nie udało się odczytać schowka',
      mediaKeys: {
        title: 'Klawisze multimedialne',
        mute: 'Wycisz',
        volumeDown: 'Ciszej',
        volumeUp: 'Głośniej',
        previous: 'Poprzedni utwór',
        playPause: 'Odtwórz lub wstrzymaj',
        next: 'Następny utwór',
        stop: 'Zatrzymaj'
      },
      pasting: {
        layout: 'Układ klawiatury hosta',
        layouts: {
          us: 'Angielski (USA)',
          uk: 'Angielski (Wielka Brytania)',
          de: 'Niemiecki',
          fr: 'Francuski',
          es: 'Hiszpański',
          it: 'Włoski',
          ptBr: 'Portugalski (Brazylia)',
          se: 'Szwedzki / fiński',
          ru: 'Rosyjski',
          ja: 'Japoński',
          ko: 'Koreański'
        },
        speed: 'Szybkość pisania',
        speeds: {
          fast: 'Szybka',
          normal: 'Normalna',
          slow: 'Wolna'
        },
        estimate: 'Czas pisania: około {{duration}}',
        untypeable: 'Znaki, których ten układ nie wpisze: {{count}}',
        untypeableAt: 'wiersz {{line}}, kolumna {{column}}',
        skipUntypeable: 'Wpisz resztę',
        shortcut: '{{shortcut}} od razu wpisuje zawartość schowka na hoście.',
        clipboardUnavailable:
          'Przeglądarka pozwala stronie czytać schowek tylko przez HTTPS. Wklej tekst do pola za pomocą Ctrl+V.',
        clipboardEmpty: 'Schowek nie zawiera tekstu.',
        tooLong: 'Tekst jest za długi. Limit to {{max}} znaków.',
        inProgress: 'Inny wklejony tekst jest już wpisywany.',
        typing: 'Wpisywanie na hoście',
        done: 'Tekst wpisany',
        canceled: 'Wklejanie anulowane',
        failed: 'Wklejanie nie powiodło się',
        cancel: 'Anuluj',
        controlBusy: 'Klawiatury używa inny kontroler.',
        hidError: 'Nie udało się wysłać naciśnięć klawiszy do hosta.'
      },
      shortcut: {
        sendFailed: 'Nie wysłano: połączenie wejścia jest przerwane',
        title: 'Skróty',
        custom: 'Niestandardowe',
        capture: 'Kliknij tutaj, aby przechwycić skrót',
        clear: 'Jasne',
        save: 'Zapisz',
        captureTips:
          'Przechwytywanie klawiszy systemowych (takich jak klawisz Windows) wymaga uprawnienia do pełnego ekranu.',
        enterFullScreen: 'Przełącz tryb pełnoekranowy.'
      },
      leaderKey: {
        saveFailed: 'Nie udało się zapisać klawisza wiodącego',
        title: 'Klawisz Leader',
        desc: 'Omiń ograniczenia przeglądarki i wyślij skróty systemowe bezpośrednio do zdalnego hosta.',
        howToUse: 'Jak używać',
        simultaneous: {
          title: 'Tryb symultaniczny',
          desc1: 'Naciśnij i przytrzymaj klawisz Leader, a następnie naciśnij skrót.',
          desc2: 'Intuicyjne, ale może kolidować ze skrótami systemowymi.'
        },
        sequential: {
          title: 'Tryb sekwencyjny',
          desc1:
            'Naciśnij klawisz Leader → naciśnij skrót w sekwencji → ponownie naciśnij klawisz Leader.',
          desc2:
            'Wymaga większej liczby kroków, ale całkowicie pozwala uniknąć konfliktów systemowych.'
        },
        enable: 'Włącz klawisz Leader',
        tip: 'Po przypisaniu jako klawisz Leader ten klawisz działa wyłącznie jako wyzwalacz skrótów i traci swoje domyślne zachowanie.',
        placeholder: 'Naciśnij klawisz Leader',
        shiftRight: 'Prawy Shift',
        ctrlRight: 'Prawy Ctrl',
        metaRight: 'Prawy Win',
        submit: 'Prześlij',
        recorder: {
          rec: 'NAGR',
          activate: 'Aktywuj klawisze',
          input: 'Proszę nacisnąć skrót...'
        }
      }
    },
    mouse: {
      jiggler: 'Poruszanie myszą',
      title: 'Mysz',
      cursor: 'Styl kursora',
      default: 'Domyślny kursor',
      pointer: 'Wskazujący kursor',
      cell: 'Kursor komórki',
      text: 'Kursor tekstowy',
      grab: 'Kursor chwytania',
      hide: 'Ukruj kursor',
      mode: 'Tryb myszki',
      absolute: 'Tryb bezwzględny',
      relative: 'Tryb względny',
      absoluteShort: 'Bezwzględny',
      relativeShort: 'Względny',
      touch: 'Tryb dotykowy',
      touchShort: 'Dotykowy',
      absoluteStalled: 'Host ignoruje mysz absolutną',
      absoluteStalledDesc:
        'Host przestał odbierać raporty myszy absolutnej, więc ruchy wskaźnika są tracone. Klawiatura działa normalnie. Zwykle pomaga odzyskanie USB; tryb względny używa innego punktu końcowego.',
      useRelative: 'Przełącz na tryb względny',
      direction: 'Kierunek kółka przewijania',
      scrollUp: 'Tak jak na tym komputerze',
      scrollDown: 'Odwrócone (naturalne przewijanie)',
      speed: 'Szybkość kółka przewijania',
      fast: 'Szybko',
      slow: 'Powoli',
      requestPointer: 'Korzystanie z trybu względnego. Kliknij pulpit, aby uzyskać wskaźnik myszy.',
      resetHid: 'Zresetuj HID',
      hidOnly: {
        switchFailed: 'Nie udało się przełączyć trybu. Sprawdź połączenie i spróbuj ponownie.',
        title: 'Tryb tylko HID',
        desc: 'Jeśli mysz i klawiatura przestaną odpowiadać, a resetowanie HID nie pomoże, może to oznaczać problem ze zgodnością między IronKVM a urządzeniem. Spróbuj włączyć tryb HID-Only, aby uzyskać lepszą kompatybilność.',
        tip1: 'Włączenie trybu HID-Only spowoduje odmontowanie wirtualnego dysku U i sieci wirtualnej',
        tip2: 'W trybie HID-Only montowanie obrazu jest wyłączone',
        rebuild: 'Zmiana trybu odbudowuje połączenie USB. IronKVM nie uruchamia się ponownie',
        enable: 'Włącz tryb HID-Only',
        disable: 'Wyłącz tryb HID-Tylko'
      },
      resetHidDone: 'Zresetowano USB HID',
      resetHidFailed: 'Nie udało się zresetować USB HID'
    },
    image: {
      delete: 'Usuń',
      inUse: 'W użyciu. Wysuń go przed usunięciem.',
      retry: 'Ponów',
      loadFailed: 'Nie udało się wczytać listy obrazów',
      readOnlyLocked: 'Wysuń dysk, aby to zmienić. Ustawienie działa przy wkładaniu obrazu.',
      title: 'Obrazy',
      loading: 'Ładowanie...',
      empty: 'Nic nie znaleziono',
      mountMode: 'Tryb montowania',
      mountFailed: 'Nie udało się zamontować obrazu',
      mountDesc:
        'W niektórych systemach wymagane jest wyjęcie dysku wirtualnego na zdalnym hoście przed zamontowaniem obrazu.',
      unmountFailed: 'Odmontowanie nie powiodło się',
      unmountDesc:
        'W niektórych systemach należy ręcznie wysunąć obraz ze zdalnego hosta przed odmontowaniem obrazu.',
      refresh: 'Odśwież listę obrazów',
      disk: 'Dysk',
      cdrom: 'CD',
      driveEmpty: 'Pusty',
      eject: 'Wysuń',
      readOnly: 'Tylko do odczytu',
      readOnlyTip: 'Dotyczy następnego obrazu włożonego do dysku.',
      noDrives: 'Brak napędów wirtualnych. Włącz dysk wirtualny w Ustawieniach.',
      insertFailed: 'Nie udało się włożyć',
      ejectFailed: 'Nie udało się wysunąć',
      insertInto: 'Włóż do napędu „{{drive}}”. Kliknij, aby zmienić.',
      loadedIn: 'W napędzie „{{drive}}”',
      attention: 'Uwaga',
      deleteConfirm: 'Czy na pewno chcesz usunąć to zdjęcie?',
      okBtn: 'Tak',
      cancelBtn: 'Nie',
      deleteFailed: 'Usuwanie nie powiodło się',
      ventoy: {
        statusNoKernel: 'Nieobsługiwane przez to oprogramowanie',
        statusNotInstalled: 'Nie zainstalowano',
        statusReady: 'Gotowe',
        statusSelected: 'Wybrane obrazy: {{count}}',
        statusInDrive: 'W napędzie dysku, {{size}}',
        noKernel:
          'Jądro tego oprogramowania nie obsługuje device-mapper, więc Ventoy nie może być używany, dopóki nie zostanie zainstalowany obraz z tą obsługą.',
        installDesc: 'Uruchamiaj hosta z kilku obrazów na jednym dysku, bez ich kopiowania.',
        install: 'Zainstaluj',
        installing: 'Pobieranie Ventoy, około 20 MB. Może to potrwać kilka minut.',
        needsData: 'Ventoy wymaga obrazu IronKVM z zamontowaną partycją /data.',
        uninstall: 'Odinstaluj',
        uninstallConfirm: 'Usunąć pliki Ventoy?',
        noImages: 'Brak obrazów do umieszczenia na dysku Ventoy.',
        onDisk: 'Na dysku Ventoy',
        missing: 'Brak pliku: {{file}}',
        remove: 'Usuń z dysku Ventoy',
        setHint:
          'Zestaw obrazów można zmieniać tylko wtedy, gdy dysk Ventoy nie jest w żadnym napędzie.',
        useAsDisk: 'Użyj jako dysku wirtualnego',
        failed: 'Żądanie Ventoy nie powiodło się',
        secureBoot:
          'Przy włączonym Secure Boot host musi raz zarejestrować klucz Ventoy w MokManager. Plik klucza ENROLL_THIS_KEY_IN_MOKMANAGER.cer znajduje się na partycji VTOYEFI.',
        readOnly:
          'Host widzi dysk jako tylko do odczytu, więc trwałość Ventoy i ventoy.json na napędzie nie działają.'
      },
      tips: {
        title: 'Jak przesłać obrazy',
        usb1: 'Podłącz urządzenie IronKVM do komputera przez USB.',
        usb2: 'Upewnij się, że dysk wirtualny jest zamontowany (Ustawienia - Dysk wirtualny).',
        usb3: 'Otwórz dysk wirtualny na swoim komputerze i skopiuj plik obrazu do katalogu głównego dysku wirtualnego.',
        scp1: 'Upewnij się że IronKVM i twój komputer są na tej samej sieci lokalnej.',
        scp2: 'Otwórz terminal na komputerze i użyj komendę SCP aby przesłać obraz do katalogu /data na IronKVM.',
        scp3: 'Przykład: scp lokalizacja-zrodlowego-obrazu root@ip-twojego-nanokvm:/data',
        tfCard: 'Karta SD',
        tf1: 'Ta metoda jest obsługiwana w systemie Linux',
        tf2: 'Usuń kartę SD od IronKVM (dla wersji FULL, rozbierz obudowę najpierw).',
        tf3: 'Włóż kartę SD do czytnika kart i podłącz do twojego komputera.',
        tf4: 'Kopjuj obraz do katalogu /data na karcie SD.',
        tf5: 'Włóż kartę SD do IronKVM.'
      }
    },
    script: {
      title: 'Skrypty',
      upload: 'Prześlij',
      run: 'Uruchom',
      runBackground: 'Uruchomiony w tle',
      runFailed: 'Uruchomienie nie powiodło się',
      attention: 'Uwaga',
      delDesc: 'Czy na pewno chcesz usunąć ten plik?',
      confirm: 'Tak',
      cancel: 'Nie',
      delete: 'Usuń',
      close: 'Zamknij',
      empty: 'Brak skryptów. Prześlij plik .sh lub .py, aby uruchomić go na płytce.',
      loadFailed: 'Nie udało się wczytać skryptów',
      uploaded: 'Przesłano skrypt',
      uploadFailed: 'Nie udało się przesłać skryptu',
      started: 'Skrypt uruchomiony w tle',
      deleteFailed: 'Nie udało się usunąć skryptu',
      waitLimit: 'Oczekiwanie na zakończenie skryptu, maksymalnie {{minutes}} minut.',
      timedOut:
        'Skrypt działał dłużej niż {{minutes}} minut i strona przestała czekać. Może nadal działać na płytce.'
    },
    terminal: {
      invalidBaud: 'Ta prędkość transmisji nie jest obsługiwana.',
      invalidPort: 'Podaj ścieżkę urządzenia w /dev, na przykład /dev/ttyS1.',
      invalidSettings: 'Nieprawidłowe ustawienia portu szeregowego. To jest powłoka samej płytki.',
      disconnected: 'Rozłączono. Naciśnij Enter, aby połączyć ponownie.',
      title: 'Terminal',
      nanokvm: 'Terminal IronKVM',
      serial: 'Terminal portu szeregowego',
      serialPort: 'Port szeregowy',
      serialPortPlaceholder: 'Wprowadź port szeregowy',
      baudrate: 'Szybkość transmisji',
      parity: 'Kontrola parzystości',
      parityNone: 'Brak kontroli',
      parityEven: 'Parzysta',
      parityOdd: 'Nieparzysta',
      flowControl: 'Kontrola przepływu',
      flowControlNone: 'Brak kontroli',
      flowControlSoft: 'Programowe',
      flowControlHard: 'Sprzętowe',
      dataBits: 'Bity danych',
      stopBits: 'Bity stopu',
      confirm: 'Ok'
    },
    wol: {
      no: 'Nie',
      yes: 'Tak',
      deleteConfirm: 'Usunąć ten zapisany adres?',
      delete: 'Usuń',
      wake: 'Wybudź',
      rename: 'Zmień nazwę',
      showMac: 'Pokaż adres MAC',
      showName: 'Pokaż nazwę',
      requestFailed: 'Nie udało się połączyć z urządzeniem, aby wysłać polecenie',
      deleteFailed: 'Nie udało się usunąć',
      renameFailed: 'Nie udało się zmienić nazwy',
      title: 'Wake-on-LAN',
      sending: 'Wysyłanie komendy...',
      sent: 'Komenda wysłana',
      input: 'Wprowadź numer adresu MAC',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'Przesyłanie nie powiodło się',
      uploadSuccess: 'Przesyłanie zakończone',
      uploading: 'Przesyłanie: {{file}}',
      downloadingPercent: 'Pobieranie ({{percent}}): {{file}}',
      downloading: 'Pobieranie: {{file}}',
      title: 'Narzędzie do pobierania obrazów',
      input: 'Proszę wprowadzić zdalny obraz URL',
      ok: 'Ok',
      disabled: '/data partycja to RO, więc nie możemy pobrać obrazu',
      uploadbox: 'Upuść plik tutaj lub kliknij, aby wybrać',
      inputfile: 'Proszę wprowadzić plik obrazu',
      NoISO: 'Brak ISO',
      sha256: 'SHA-256 (opcjonalnie)',
      sha256Placeholder: 'Wprowadź 64-znakową sumę kontrolną SHA-256',
      invalidSHA256: 'SHA-256 musi być 64-znakowym ciągiem szesnastkowym',
      failed: 'Pobieranie nie powiodło się',
      success: 'Pobieranie zakończone pomyślnie',
      checksumFailed: 'Pobieranie nie powiodło się: weryfikacja SHA-256 nie powiodła się',
      cancel: 'Anuluj',
      cancelFailed: 'Nie udało się anulować pobierania',
      bootMenu: 'Menu rozruchowe (netboot.xyz)',
      bootMenuPresent: '{{file}} jest już na urządzeniu, z poprawną sumą kontrolną',
      bootMenuDesc: 'Pobierz obraz ISO netboot.xyz ze sprawdzoną sumą kontrolną do wirtualnego CD'
    },
    power: {
      title: 'Zasilanie',
      showConfirm: 'Potwierdzenie',
      showConfirmTip:
        'Pytaj przed krótkim naciśnięciem zasilania. Reset i długie naciśnięcie pytają zawsze.',
      reset: 'Resetuj',
      power: 'Zasilanie',
      powerShort: 'Zasilanie (krótkie kliknięcie)',
      powerLong: 'Zasilanie (długie kliknięcie)',
      resetConfirm: 'Kontynuować operację resetowania?',
      powerConfirm: 'Kontynuować zasilanie?',
      okBtn: 'Tak',
      cancelBtn: 'Nie',
      hostOs: 'System hosta',
      hostOsTip: 'Wysyłane jako klawisze USB. O ich działaniu decyduje host.',
      sleep: 'Uśpij',
      wake: 'Wybudź',
      wakeKey: 'Wybudź klawiszem Shift',
      powerDown: 'Wyłącz',
      sleepConfirm: 'Uśpić hosta?',
      powerDownConfirm: 'Wysłać do hosta klawisz wyłączenia?',
      wakeTip:
        'Uśpiony host często ignoruje Wybudź od urządzenia, które go uśpiło. Wybudź klawiszem Shift naciska klawisz na klawiaturze, który akceptuje więcej hostów.',
      led: 'Dioda zasilania',
      ledOn: 'Świeci',
      ledOff: 'Nie świeci',
      ledUnknown: 'Nieznany',
      ledConnected: 'Dioda zasilania podłączona',
      ledConnectedTip:
        'Włącz tylko wtedy, gdy złącze diody zasilania hosta jest podłączone do płytki. Bez niego stan zasilania jest nieznany.',
      ledConnectedFailed: 'Nie udało się zapisać ustawienia diody zasilania',
      powerLongConfirm:
        'Przytrzymać przycisk zasilania przez {{seconds}} s? To odcina zasilanie bez zamknięcia systemu.',
      done: 'Przycisk naciśnięty',
      failed: 'Nie udało się nacisnąć przycisku'
    },
    settings: {
      title: 'Ustawienia',
      nav: {
        system: 'System',
        network: 'Sieć',
        access: 'Dostęp',
        integrations: 'Integracje',
        boot: 'Rozruch',
        browser: 'Ta przeglądarka',
        search: 'Znajdź ustawienie',
        noMatch: 'Brak pasujących ustawień',
        locked: 'Trwa operacja. Inne strony i zamknięcie są niedostępne do jej zakończenia.',
        vpnProvider: 'Dostawca VPN'
      },
      mcp: {
        keyNote:
          'MCP używa własnego klucza API, pokazanego poniżej. Klucze ze strony Klucze API tu nie działają.',
        title: 'Usługa MCP',
        service: 'Zdalne sterowanie MCP',
        serviceDesc:
          'Zezwalaj zaufanym klientom MCP na sterowanie klawiaturą i myszą oraz wykonywanie zrzutów ekranu',
        securityWarning:
          'Każdy, kto ma ten klucz API, może sterować zdalnym hostem i wyświetlać jego ekran. Używaj HTTPS i włączaj usługę tylko w zaufanych sieciach.',
        endpoint: 'Punkt końcowy',
        apiKey: 'Klucz API',
        regenerateConfirmTitle: 'Wygenerować ponownie klucz API MCP?',
        regenerateConfirmDesc: 'Bieżący klucz natychmiast przestanie działać.',
        enableConfirmTitle: 'Włączyć zewnętrzne sterowanie MCP?',
        enableConfirmDesc:
          'Włączenie MCP zatrzyma PicoClaw i zamknie wszystkie aktywne sesje PicoClaw.',
        failed: 'Operacja MCP nie powiodła się',
        copyFailed: 'Kopiowanie nie powiodło się. Skopiuj ręcznie.',
        okBtn: 'Potwierdź',
        cancelBtn: 'Anuluj',
        showKey: 'Pokaż klucz',
        hideKey: 'Ukryj klucz',
        regenerateKey: 'Wygeneruj nowy klucz'
      },
      redfish: {
        example: 'Przykład',
        title: 'Redfish',
        service: 'Usługa Redfish',
        serviceDesc:
          'API DMTF Redfish do sterowania zasilaniem, nośnikami wirtualnymi i odczytu stanu z narzędzi takich jak redfishtool i Ansible. Wyłączenie kończy wszystkie sesje Redfish.',
        endpoint: 'Katalog główny usługi',
        httpsOn: 'Płytka działa przez HTTPS, którego wymaga większość narzędzi Redfish.',
        httpsOff:
          'Płytka działa przez zwykłe HTTP. Większość narzędzi Redfish wymaga HTTPS: włącz je w „Ustawienia > Sieć”.',
        credentials:
          'Redfish akceptuje konta KVM z uwierzytelnianiem Basic lub sesją Redfish oraz klucze API wysyłane jako X-Auth-Token. Kluczami API zarządza się na stronie „Klucze API”.',
        powerActions: 'Akcje zasilania',
        powerActionsDesc:
          'Obecnie dostępne typy resetu. On, ForceOff i GracefulShutdown wymagają znajomości stanu zasilania, więc są dostępne tylko wtedy, gdy w menu zasilania włączono „Dioda zasilania podłączona”.',
        sessions: 'Sesje',
        noSessions: 'Brak otwartych sesji Redfish',
        created: 'Utworzono',
        lastUsed: 'Ostatnie użycie',
        refresh: 'Odśwież',
        end: 'Zakończ',
        endConfirmTitle: 'Zakończyć tę sesję Redfish?',
        endConfirmDesc:
          'Jej token natychmiast przestanie działać. Klient będzie musiał zalogować się ponownie.',
        failed: 'Operacja Redfish nie powiodła się',
        copyFailed: 'Kopiowanie nie powiodło się. Skopiuj ręcznie.',
        okBtn: 'Potwierdź',
        cancelBtn: 'Anuluj'
      },
      ipmi: {
        copyBeforeSave: 'Skopiuj hasło teraz. Po zapisaniu nie da się go ponownie wyświetlić.',
        noLogin:
          'IPMI jest włączone, ale żadne aktywne konto nie ma hasła IPMI, więc nikt się nie zaloguje. Ustaw je poniżej.',
        title: 'IPMI',
        warning:
          'Uwierzytelnianie IPMI jest z założenia słabe. Każdy, kto ma dostęp do płytki i zna nazwę użytkownika, może pobrać skrót hasła IPMI tego użytkownika i próbować złamać go offline. Używaj generowanych haseł, włączaj IPMI tylko w zaufanej sieci i tam, gdzie narzędzie to obsługuje, wybieraj Redfish przez HTTPS.',
        service: 'IPMI przez LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) na porcie UDP 623 do zasilania i stanu hosta. IPMI 1.5 i zestaw szyfrów 0 są odrzucane. Wyłączenie kończy wszystkie sesje IPMI.',
        example: 'Przykład',
        copyFailed: 'Kopiowanie nie powiodło się. Skopiuj ręcznie.',
        ledOn: 'Dostępne są stan zasilania, on, off, soft, cycle i reset.',
        ledOff:
          '„Dioda zasilania podłączona” jest wyłączona w menu zasilania, więc stan zasilania jest nieznany. Działa tylko „power reset”: status, on, off, soft i cycle są odrzucane.',
        accounts: 'Konta',
        accountsDesc:
          'IPMI loguje się kontami KVM, każde z własnym hasłem IPMI, odrębnym od hasła webowego. Administratorzy dostają ADMINISTRATOR. Użytkownicy dostają USER: z „-L USER” mogą odczytać stan zasilania, ale nie mogą go zmienić.',
        passwordSet: 'Hasło IPMI ustawione',
        passwordNotSet: 'Brak hasła IPMI: logowanie przez IPMI niemożliwe',
        nameTooLong: 'Nazwa ma więcej niż 16 znaków, na co IPMI nie pozwala',
        accountDisabled: 'Konto jest wyłączone',
        setPassword: 'Ustaw hasło',
        changePassword: 'Zmień hasło',
        remove: 'Usuń',
        removeConfirmTitle: 'Usunąć hasło IPMI konta {{user}}?',
        removeConfirmDesc:
          'Konto nie będzie mogło logować się przez IPMI, a jego sesje IPMI zostaną zakończone.',
        passwordTitle: 'Hasło IPMI dla {{user}}',
        passwordDesc:
          'Od 12 do 20 drukowalnych znaków ASCII, inne niż hasło webowe. IPMI wymaga, by płytka przechowywała hasło w postaci, którą może odczytać, więc użyj hasła, którego nie używasz nigdzie indziej. Skopiuj je przed zapisaniem: nie zostanie pokazane ponownie.',
        passwordPlaceholder: 'Hasło IPMI',
        generate: 'Generuj',
        copy: 'Kopiuj',
        save: 'Zapisz',
        passwordLength: 'Użyj od 12 do 20 znaków.',
        passwordChars: 'Używaj tylko drukowalnych znaków ASCII.',
        saved: 'Hasło IPMI zapisane',
        failed: 'Operacja IPMI nie powiodła się',
        okBtn: 'Potwierdź',
        cancelBtn: 'Anuluj'
      },
      ssh: {
        service: 'Serwer SSH',
        serviceDesc: 'Uruchom sshd teraz i przy każdym starcie',
        failed: 'Nie udało się wczytać ustawień SSH',
        rootDefault: 'root nadal ma hasło fabryczne',
        rootEmpty: 'root nie ma hasła',
        rootWarning:
          'Każdy, kto dotrze do konsoli lub SSH, może zalogować się jako root. Ustaw hasło w {{account}} > {{password}}: dla właściciela urządzenia ustawia ono też hasło roota.',
        connection: 'Połączenie',
        command: 'Zaloguj się jako root',
        port: 'Port',
        viaVpn: 'Przez {{name}}',
        notRunning: 'sshd nie działa. Włącz serwer SSH, aby się połączyć.',
        hostKeys: 'Odciski kluczy hosta',
        hostKeysDesc: 'Porównaj je z tym, co ssh pokaże przy pierwszym połączeniu.',
        noHostKeys: 'Brak kluczy hosta. sshd tworzy je przy pierwszym uruchomieniu.',
        keys: 'Autoryzowane klucze',
        keysDesc:
          'Klucze publiczne, które mogą logować się jako root. Są przechowywane na partycji danych, więc aktualizacje je zachowują.',
        noKeys: 'Brak autoryzowanych kluczy.',
        noComment: 'bez komentarza',
        addPlaceholder: 'Wklej jeden klucz publiczny, np. zawartość ~/.ssh/id_ed25519.pub',
        add: 'Dodaj klucz',
        added: 'Klucz dodany',
        removed: 'Klucz usunięty',
        deleteConfirm: 'Usunąć ten klucz?',
        deleteConfirmDesc: 'Nie będzie mógł się już zalogować. Otwarte sesje pozostają otwarte.',
        invalidKey: 'To nie jest klucz publiczny. Wklej jeden wiersz z pliku .pub.',
        keyOptions: 'Klucze z opcjami, takimi jak command= lub from=, nie są tu przyjmowane.',
        duplicateKey: 'Ten klucz jest już autoryzowany.',
        lastKey: 'Ostatniego klucza nie można usunąć, gdy włączone jest logowanie tylko kluczem.',
        keysOnly: 'Tylko klucze',
        keysOnlyDesc:
          'Wyłącz logowanie hasłem i keyboard-interactive. Otwarte sesje pozostają otwarte.',
        keysOnlyNeedsKey:
          'Najpierw dodaj autoryzowany klucz, inaczej nikt nie mógłby się zalogować.',
        keysOnlyOn: 'Logowanie hasłem wyłączone',
        keysOnlyOff: 'Logowanie hasłem włączone',
        notHonoured:
          'sshd w tym obrazie nie czyta tego ustawienia, więc logowanie hasłem pozostaje włączone.',
        reloadFailed:
          'Zapisano, ale nie udało się przeładować sshd. Zadziała przy następnym starcie sshd.',
        notApplied:
          'sshd nadal przyjmuje hasła. Wyłącz i włącz serwer SSH, aby zastosować ustawienie.'
      },
      vnc: {
        address: 'Adres',
        certHint:
          'VeNCrypt X509Plain używa samopodpisanego certyfikatu urządzenia, więc klient ostrzega przy pierwszym połączeniu. Zaakceptuj go albo zapisz certyfikat z adresu HTTPS tej strony i przekaż go TigerVNC opcją -X509CA=<plik>.',
        example: 'Przykład',
        title: 'VNC',
        service: 'Serwer VNC',
        serviceDesc:
          'Pozwala klientowi VNC, takiemu jak TigerVNC lub Remmina, wyświetlać i sterować hostem. Klient musi obsługiwać kodowanie Tight. Jedna sesja naraz.',
        credentials:
          'Zaloguj się kontem KVM. Połączenie jest szyfrowane certyfikatem TLS płytki (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'Port TCP, na którym nasłuchuje serwer.',
        maxFps: 'Limit klatek',
        maxFpsDesc: 'Największa liczba klatek na sekundę wysyłana do klienta.',
        vncAuth: 'Proste uwierzytelnianie VNC',
        vncAuthDesc: 'Dla klientów bez VeNCrypt. Sprawdza osobne hasło VNC zamiast konta.',
        vncAuthWarning:
          'Proste uwierzytelnianie VNC nie szyfruje połączenia. Każdy na ścieżce sieciowej może zobaczyć ekran i naciśnięcia klawiszy. Używaj go tylko w zaufanej sieci.',
        password: 'Hasło VNC',
        passwordSet: 'Hasło jest ustawione. Wpisz nowe, aby je zmienić.',
        passwordInvalid: 'Hasło VNC musi mieć od 6 do 8 znaków.',
        save: 'Zapisz',
        saved: 'Ustawienia zapisane',
        state: 'Stan',
        listening: 'Nasłuchuje na porcie {{port}}',
        notListening: 'Nie nasłuchuje',
        noSession: 'Brak otwartej sesji',
        client: 'Klient',
        user: 'Użytkownik',
        method: 'Uwierzytelnianie',
        methodVencrypt: 'Konto przez TLS',
        methodVnc: 'Hasło VNC',
        since: 'Połączono od',
        resolution: 'Rozdzielczość',
        framesSent: 'Wysłane klatki',
        lastError: 'Ostatnia sesja zakończyła się: {{error}}',
        refresh: 'Odśwież',
        disconnect: 'Rozłącz',
        disconnectConfirmTitle: 'Zakończyć sesję VNC?',
        disconnectConfirmDesc:
          'Klient zostanie od razu rozłączony, a wszystkie przytrzymane klawisze i przyciski zostaną zwolnione.',
        failed: 'Operacja VNC nie powiodła się',
        okBtn: 'Potwierdź',
        cancelBtn: 'Anuluj'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Watchdog hosta',
        serviceDesc:
          'Jeśli host powinien działać, a jego obraz nie zmienia się lub brak sygnału HDMI przez czas limitu, płytka naciska reset albo wyłącza i ponownie włącza hosta.',
        stillWarning:
          'Host, którego ekran przechodzi w uśpienie lub którego obraz stoi w miejscu podczas pracy, wygląda na zawieszony. Wyłącz usypianie ekranu na hoście albo ustaw adres do pingowania.',
        ledHint:
          'Przełącznik "Dioda zasilania podłączona" w menu zasilania jest wyłączony. Watchdog nie widzi, kiedy host jest wyłączony, więc traktuje go jako zawsze włączony.',
        timeout: 'Limit czasu',
        timeoutDesc: 'Jak długo host może nie dawać oznak życia, zanim watchdog zadziała.',
        action: 'Akcja',
        actionDesc:
          'Wyłączenie i włączenie przytrzymuje przycisk zasilania przez 5 sekund, a potem naciska go ponownie.',
        actionReset: 'Resetuj',
        actionPower: 'Wyłącz i włącz',
        cooldown: 'Przerwa',
        cooldownDesc: 'Najkrótszy czas między dwiema akcjami.',
        maxPerHour: 'Akcje na godzinę',
        maxPerHourDesc: 'Największa liczba akcji w ciągu godziny.',
        pingHost: 'Adres do pingowania',
        pingHostDesc:
          'Adres IP hosta. Odpowiedź liczy się jako oznaka życia. Pozostaw puste, aby nie pingować.',
        pingHostInvalid: 'Wpisz adres IPv4 lub IPv6.',
        minutes: 'min',
        save: 'Zapisz',
        saved: 'Zapisano',
        state: 'Detektor',
        status: {
          off: 'Wyłączony',
          watching: 'Obserwuje',
          hostOff: 'Host wyłączony',
          captureOff: 'Przechwytywanie HDMI wyłączone',
          cooldown: 'Przerwa',
          capped: 'Osiągnięto limit godzinowy',
          acting: 'Działa'
        },
        signal: 'Sygnał HDMI',
        yes: 'Tak',
        no: 'Nie',
        led: 'Dioda zasilania',
        on: 'Świeci',
        off: 'Nie świeci',
        ledNotConnected: 'Niepodłączona',
        ping: 'Ping',
        pingNotSet: 'Nie ustawiono',
        pingReply: 'Odpowiada',
        pingNoReply: 'Brak odpowiedzi',
        lastChange: 'Ostatnia zmiana obrazu',
        never: 'Nigdy',
        actsIn: 'Zadziała za',
        actionsLastHour: 'Akcje w ostatniej godzinie',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Dziennik',
        noLog: 'Watchdog jeszcze nie zadziałał.',
        refresh: 'Odśwież',
        reasonFrozen: 'Obraz się nie zmienił',
        reasonNoSignal: 'Brak sygnału HDMI',
        stuckFor: 'brak oznak życia przez {{duration}}',
        pressFailed: 'Naciśnięcie nie powiodło się: {{error}}',
        noScreenshot: 'Brak zrzutu ekranu',
        failed: 'Operacja watchdoga nie powiodła się',
        powerNeedsLed: 'Cykl zasilania wymaga opcji „Dioda zasilania podłączona” w menu zasilania.',
        noLedConfirmTitle: 'Włączyć watchdog bez diody zasilania?',
        noLedConfirmDesc:
          'Płytka nie widzi, kiedy host jest wyłączony, więc traktuje go jako zawsze włączony. Jeśli wyłączysz hosta, watchdog po upływie limitu czasu naciśnie reset. Podłącz diodę zasilania, aby tego uniknąć.',
        noLedConfirmOk: 'Włącz',
        cancel: 'Anuluj'
      },
      netboot: {
        title: 'Rozruch sieciowy',
        description:
          'Uruchom hosta z sieci: iPXE i menu obrazów z KVM przez sieciowe łącze USB albo netboot.xyz przez proxy DHCP w sieci LAN.',
        addon: 'dnsmasq i pliki rozruchowe',
        addonDesc:
          'Instalowane na /data: dnsmasq z Alpine, iPXE i netboot.xyz z ich wydań, każdy sprawdzony sumą kontrolną.',
        install: 'Zainstaluj',
        installing: 'Instalowanie. Może to potrwać kilka minut.',
        uninstall: 'Odinstaluj',
        uninstallConfirm: 'Wyłączyć rozruch sieciowy i usunąć dnsmasq oraz pliki rozruchowe?',
        needsData: 'Rozruch sieciowy wymaga obrazu IronKVM z zamontowaną partycją /data.',
        usb: 'Na sieciowym łączu USB',
        usbDesc:
          'Gdy sieciowe łącze USB jest włączone, obsługuje je dnsmasq zamiast udhcpd. Host dostaje swój jedyny adres bez routera i bez serwera DNS, iPXE dla swojej architektury oraz menu obrazów ISO z KVM.',
        linkOff: 'Sieciowe łącze USB jest wyłączone. Włącz je w Urządzenie, Sieć USB.',
        menuUrl: 'Menu',
        leases: 'Dzierżawa hosta',
        noLeases: 'Jeszcze brak',
        netbootxyzNote:
          'netboot.xyz w menu ładuje się z internetu, do którego łącze USB nie sięga. Host potrzebuje internetu na innym porcie sieciowym.',
        lan: 'Proxy DHCP w sieci LAN',
        lanDesc:
          'Odpowiada klientom PXE w sieci LAN, podając netboot.xyz, który potem ładuje swoje menu z internetu. Nigdy nie przydziela adresów i nie udostępnia obrazów z KVM.',
        lanWarning:
          'netboot.xyz jest oferowany każdemu klientowi PXE w tej sieci LAN, nie tylko hostowi. Włącz to tylko w sieci, którą kontrolujesz.',
        lanConfirm: 'Włączyć proxy DHCP w sieci LAN?',
        lanInterface: 'LAN',
        running: 'Działa',
        stopped: 'Nie działa',
        images: 'Obrazy w menu',
        noImages: 'Brak obrazów ISO w katalogu obrazów.',
        boots: 'Ostatnie rozruchy',
        noBoots: 'Host jeszcze niczego nie pobrał.',
        log: 'Dziennik dnsmasq',
        refresh: 'Odśwież',
        okBtn: 'Potwierdź',
        cancelBtn: 'Anuluj',
        failed: 'Operacja rozruchu sieciowego nie powiodła się'
      },
      about: {
        title: 'IronKVM - informacje',
        information: 'Informacje o systemie',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Wersja oprogramowania',
        applicationTip: 'Wersja aplikacji web IronKVM',
        image: 'Wersja obrazu',
        imageTip: 'Obraz karty IronKVM i obraz systemu NanoKVM, na którym jest zbudowany',
        kernel: 'Wersja jądra',
        kernelTip: 'Wydanie aktualnie działającego jądra Linux',
        deviceKey: 'Klucz urządzenia',
        videoMemory: 'Pamięć wideo',
        videoMemoryTip:
          'Pamięć zarezerwowana na przechwytywanie wideo. Nie jest współdzielona z resztą systemu.',
        videoMemoryGenerations_one: '{{count}} wcześniejsza sesja IronKVM zajmuje pamięć wideo',
        videoMemoryGenerations_few: '{{count}} wcześniejsze sesje IronKVM zajmują pamięć wideo',
        videoMemoryGenerations_many: '{{count}} wcześniejszych sesji IronKVM zajmuje pamięć wideo',
        videoMemoryGenerations_other: '{{count}} wcześniejszej sesji IronKVM zajmuje pamięć wideo',
        videoMemoryReboot: 'Uruchom ponownie, aby ją odzyskać.',
        community: 'Społeczność',
        hostname: 'Nazwa hosta',
        hostnameUpdated: 'Zaktualizowano nazwę hosta. Uruchom ponownie, aby zastosować.',
        ipType: {
          Wired: 'Przewodowy',
          Wireless: 'Bezprzewodowe',
          Other: 'Inne'
        },
        hostnameInvalid:
          'Użyj liter, cyfr i łączników, do 63 w każdej części oddzielonej kropką. Bez łącznika na początku ani na końcu części.',
        hostnameFailed: 'Nie udało się zmienić nazwy hosta',
        editHostname: 'Edytuj nazwę hosta',
        docs: 'Dokumentacja',
        hardware: 'Sprzęt',
        hardwareFaq: 'FAQ sprzętu',
        disclaimer:
          'IronKVM: wzmocnione firmware społeczności dla Sipeed NanoKVM. Niezwiązane z firmą Sipeed.',
        basedOn: 'na bazie NanoKVM {{version}}'
      },
      preferences: {
        title: 'Preferencje'
      },
      performance: {
        title: 'Wydajność'
      },
      appearance: {
        thisBrowser: 'Ta przeglądarka',
        thisBrowserDesc:
          'Zapisywane tylko w tej przeglądarce. Inne przeglądarki mają własne ustawienia.',
        deviceWide: 'Urządzenie',
        deviceWideDesc: 'Zapisywane na urządzeniu. Dotyczy każdego, kto je otworzy.',
        language: 'Język',
        languageDesc: 'Wybierz język interfejsu',
        webTitle: 'Tytuł strony internetowej',
        webTitleDesc: 'Dostosuj tytuł strony internetowej',
        menuBar: {
          title: 'Pasek menu',
          mode: 'Tryb wyświetlania',
          modeDesc: 'Wyświetl pasek menu na ekranie',
          modeOff: 'Wyłączone',
          modeAuto: 'Automatyczne ukrywanie',
          modeAlways: 'Zawsze widoczny',
          keyboardLedStatus: 'Wskaźniki blokad klawiatury',
          keyboardLedStatusDesc:
            'Wyświetl stan Num Lock, Caps Lock i Scroll Lock zdalnego komputera',
          icons: 'Ikony podmenu',
          iconsDesc: 'Wyświetla ikony podmenu na pasku menu'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Stan blokad zdalnej klawiatury',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Włączone',
        off: 'Wyłączone',
        unknown: 'Nieznany'
      },
      device: {
        title: 'Urządzenie',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'Jasność OLED',
          brightnessDescription: 'Niższy poziom wydłuża żywotność wyświetlacza',
          brightnessLevels: {
            '64': 'Najniższa',
            '96': 'Niska',
            '128': 'Średnia',
            '160': 'Wysoka',
            '207': 'Domyślna',
            '255': 'Maksymalna'
          },
          0: 'Nigdy',
          15: '15 s',
          30: '30 s',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 godzina'
        },
        sections: {
          video: 'Wideo',
          usb: 'USB',
          frontPanel: 'Panel przedni'
        },
        hidModeDesc:
          'Jeśli host nie przyjmuje klawiatury i myszy, spróbuj trybu tylko HID. Wyłącza on napędy wirtualne i sieć.',
        resetHidDesc:
          'Ponownie podłącza klawiaturę i mysz do hosta. Użyj, gdy wprowadzanie przestanie działać.',
        cpuFreq: {
          title: 'Częstotliwość CPU',
          description: 'Ustaw taktowanie CPU stosowane przy następnym uruchomieniu',
          tip: 'CPU startuje z częstotliwością 850 MHz i jest przystosowany do 1000 MHz. Nowa wartość jest stosowana przy następnym uruchomieniu, a nie w trakcie działania systemu. 1000 MHz mieści się w specyfikacji; przy obu ustawieniach temperatura jest daleko od limitu.',
          running: 'Obecnie: {{mhz}} MHz',
          rebootToApply: 'uruchom ponownie, aby zastosować',
          rebootConfirm: 'Uruchomić ponownie teraz, aby zastosować {{mhz}} MHz?'
        },
        swap: {
          title: 'Zamień',
          disable: 'Wyłącz',
          description: 'Ustaw rozmiar pliku wymiany',
          tip: 'Włączenie tej funkcji może skrócić żywotność karty SD!'
        },
        zram: {
          title: 'Skompresowana pamięć wymiany (zram)',
          description: 'Pamięć wymiany w skompresowanej RAM zamiast na karcie SD',
          tip: 'zram trzyma pamięć wymiany poza kartą SD, więc jej nie zużywa. Nie ma za nim wymiany na dysku: gdy zram się zapełni, jądro zatrzyma proces zamiast powoli stronicować. Limit pamięci określa, ile RAM może zająć zram.',
          unavailable: 'Moduły jądra nie są zainstalowane na tym urządzeniu',
          inactive: 'Włączone, ale urządzenie się nie uruchomiło',
          active: 'Aktywne - {{used}} z {{total}}, {{ratio}}x',
          off: 'Wyłączone',
          detail: {
            algorithm: 'Algorytm: {{algorithm}}',
            memory: 'Użyta pamięć: {{used}} z {{limit}}',
            memoryNoLimit: 'Użyta pamięć: {{used}}, bez limitu',
            counters:
              'Strony wczytane: {{in}}, zapisane: {{out}} (wszystkie urządzenia wymiany, od uruchomienia)'
          }
        },
        mouseJiggler: {
          title: 'Jiggler myszy',
          description: 'Uniemożliwia uśpienie zdalnego hosta',
          disable: 'Wyłącz',
          absolute: 'Tryb absolutny',
          relative: 'Tryb względny'
        },
        mdns: {
          description: 'Włącz usługę wykrywania mDNS',
          tip: 'Wyłączanie, jeśli nie jest potrzebne'
        },
        hdmi: {
          description: 'Włącz HDMI/wyjście monitora',
          idleTimeoutTitle: 'Limit czasu bezczynności przechwytywania',
          idleTimeoutDescription: 'Zatrzymaj przechwytywanie HDMI po czasie bez aktywnych widzów:',
          minutes: 'min'
        },
        hidOnly: 'HID – tylko tryb',
        hidOnlyDesc:
          'Przestań emulować urządzenia wirtualne, zachowując jedynie podstawową kontrolę HID',
        disk: 'Dysk wirtualny',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Sieć wirtualna',
        networkDesc: 'Zamontuj wirtualną kartę sieciową na zdalnym hoście',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Host:',
          description:
            'Prywatne połączenie sieciowe ze zdalnym hostem przez kabel USB. Host otrzymuje adres bez bramy i bez DNS, więc nie może dotrzeć do Twojej sieci LAN przez IronKVM.',
          off: 'Wyłączone',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (dla hostów bez NCM)',
          rndis: 'RNDIS (już niedostępne)',
          rndisNote:
            'To połączenie używa RNDIS, które nie jest już oferowane. Wybierz NCM lub ECM.',
          subnet: 'Podsieć',
          subnetDesc:
            'Prywatna sieć IPv4, od /24 do /30. IronKVM zajmuje pierwszy adres, host drugi.',
          invalidSubnet: 'Wpisz podsieć, na przykład 172.31.255.0/30.',
          apply: 'Zastosuj',
          confirm: 'Połączyć ponownie urządzenie USB?',
          reenumerate:
            'Zastosowanie odbudowuje połączenie USB. Host na kilka sekund traci klawiaturę, mysz i dysk wirtualny.'
        },
        audio: 'Wirtualny głośnik',
        audioDesc:
          'Udostępnij zdalnemu hostowi kartę dźwiękową USB, aby go słyszeć. Host musi wybrać ją jako urządzenie wyjściowe. Przełączenie odbudowuje połączenie USB.',
        audioNote: 'Dźwięk jest dostępny w obu trybach H.264 (WebRTC i Direct), ale nie w MJPEG',
        console: 'Konsola szeregowa',
        consoleDesc:
          'Udostępnij zdalnemu hostowi port szeregowy USB do logowania do tego IronKVM, gdy sieć jest niedostępna',
        consoleTip:
          'Każdy, kto kontroluje zdalny host, zobaczy monit logowania do tego IronKVM. Ustaw silne hasło przed włączeniem (Konto - Zmień hasło).',
        endpoints: {
          title: 'Punkty końcowe USB',
          used: 'Użyto {{used}} z {{total}}',
          cost: 'używa {{cost}}',
          needs: 'wymaga {{cost}}',
          full: 'Za mało punktów końcowych USB. Najpierw wyłącz coś innego.',
          inactive:
            'Włączone, ale nie działa: kontrolerowi USB zabrakło punktów końcowych. Wyłącz inne urządzenie, a to uruchomi się od razu.',
          explain:
            'Kontroler USB ma stałą liczbę wejściowych punktów końcowych i to one są tu liczone. Jeśli włączono więcej urządzeń, niż się mieści, klawiatura i mysz zostają, a reszta jest wyłączana.',
          error: 'Nie udało się połączyć z urządzeniem. Spróbuj ponownie.',
          fitTogether: 'Razem mieszczą się: {{sets}}'
        },
        reboot: 'Uruchom ponownie',
        rebootDesc: 'Czy na pewno chcesz ponownie uruchomić IronKVM?',
        okBtn: 'Tak',
        cancelBtn: 'Nie',
        rebootFailed: 'Ponowne uruchomienie nie powiodło się'
      },
      network: {
        title: 'Sieć',
        wifi: {
          disconnectBtn: 'Rozłącz',
          disconnectWarning:
            'Jeśli łączysz się z IronKVM przez tę sieć Wi-Fi, ta strona straci połączenie.',
          disconnected: 'Wi-Fi rozłączone',
          title: 'Wi-Fi',
          description: 'Skonfiguruj Wi-Fi',
          apMode: 'Tryb AP jest włączony, połącz z Wi-Fi skanując kod QR',
          connect: 'Połącz Wi-Fi',
          connectDesc1: 'Wprowadź SSID sieci i hasło',
          connectDesc2: 'Wprowadź hasło, aby połączyć się z tą siecią',
          disconnect: 'Czy na pewno chcesz rozłączyć sieć?',
          failed: 'Połączenie nie powiodło się, spróbuj ponownie.',
          ssid: 'Nazwa',
          password: 'Hasło',
          joinBtn: 'Połącz',
          confirmBtn: 'OK',
          cancelBtn: 'Anuluj'
        },
        tls: {
          description: 'Włącz protokół HTTPS',
          tip: 'Uwaga: użycie HTTPS może zwiększyć opóźnienie, szczególnie w trybie wideo MJPEG.',
          restarting: 'Ponowne uruchamianie serwera urządzenia, potrwa to około dwóch minut...',
          waiting: 'Oczekiwanie na odpowiedź urządzenia...',
          waitingHttp: 'Powrót do http. Jeśli strona nie otworzy się sama, odśwież ją.',
          failed: 'Nie udało się zmienić ustawienia HTTPS',
          enableConfirm: 'Włączyć HTTPS?',
          disableConfirm: 'Wyłączyć HTTPS?',
          confirmDesc:
            'Spowoduje to wylogowanie i ponowne uruchomienie serwera urządzenia, co trwa około dwóch minut. Następnie strona otworzy {{url}}.',
          confirmOk: 'Kontynuuj',
          confirmCancel: 'Anuluj'
        },
        ethernet: {
          title: 'Adres IP',
          description: 'Skonfiguruj, w jaki sposób IronKVM otrzymuje adres w sieci przewodowej',
          dhcp: 'DHCP',
          manual: 'Ręcznie',
          networkDetails: 'Szczegóły sieci',
          interface: 'Interfejs',
          ipAddress: 'Adres IP',
          subnetMask: 'Maska podsieci',
          router: 'Router',
          save: 'Zastosuj',
          invalidAddress: 'Wprowadź prawidłowy adres IP',
          invalidMask: 'Wprowadź prawidłową maskę podsieci, na przykład 255.255.255.0 lub 24',
          invalidRouter: 'Wprowadź prawidłowy adres routera',
          addressRequired: 'Adres IP jest wymagany',
          maskRequired: 'Maska podsieci jest wymagana',
          applyTitle: 'Zmienić adres IronKVM?',
          applyWarning:
            'Połączenie z tą stroną zostanie utracone. IronKVM zastosuje nowy adres i czeka {{seconds}} sekund, aż dotrzesz do niego pod tym adresem. Dotarcie do niego zachowa zmianę. Jeśli nic do niego nie dotrze, IronKVM przywróci poprzednie ustawienia.',
          applyConfirm: 'Zastosuj',
          applyCancel: 'Anuluj',
          applyFailed: 'Nie udało się zastosować adresu',
          trialTitle: 'Oczekiwanie na potwierdzenie',
          trialDhcp: 'IronKVM prosi o adres przez DHCP.',
          trialStatic: 'IronKVM jest teraz pod adresem {{address}}.',
          trialInstruction:
            'Otwórz IronKVM pod jego nowym adresem i zaloguj się, jeśli o to poprosi. Dotarcie do niego zachowa zmianę. Jeśli nic nie dotrze do IronKVM w ciągu {{seconds}} sekund, przywróci on poprzednie ustawienia.',
          trialOpen: 'Otwórz nowy adres',
          trialKeep: 'Zachowaj te ustawienia',
          trialKept: 'Nowy adres został zapisany',
          trialKeepFailed: 'Nie udało się zachować ustawień',
          trialGone: 'Zmiana została już cofnięta. Spróbuj ponownie.',
          unsaved: 'Niezapisane zmiany'
        },
        dns: {
          title: 'DNS',
          description: 'Skonfiguruj serwery DNS dla IronKVM',
          mode: 'Tryb',
          dhcp: 'DHCP',
          manual: 'Ręcznie',
          add: 'Dodaj DNS',
          save: 'Zapisz',
          invalid: 'Wprowadź prawidłowy adres IP',
          noDhcp: 'Brak obecnie dostępnego DNS z DHCP',
          saved: 'Ustawienia DNS zapisane',
          saveFailed: 'Nie udało się zapisać ustawień DNS',
          unsaved: 'Niezapisane zmiany',
          maxServers: 'Dozwolone jest maksymalnie {{count}} serwerów DNS',
          dnsServers: 'Serwery DNS',
          dhcpServersDescription: 'Serwery DNS są automatycznie pobierane z DHCP',
          manualServersDescription: 'Serwery DNS można edytować ręcznie',
          networkDetails: 'Szczegóły sieci',
          interface: 'Interfejs',
          ipAddress: 'Adres IP',
          subnetMask: 'Maska podsieci',
          router: 'Router',
          none: 'Brak'
        }
      },
      vpn: {
        connect: 'Połącz',
        connectDesc: 'Dołącz do sieci {{name}}. Wyłączenie rozłącza bez zatrzymywania usługi.',
        kvmUrl: 'Adres KVM',
        moreTip: 'Więcej działań',
        restartTip: 'Uruchom ponownie',
        stopTip: 'Zatrzymaj',
        updateTip: 'Aktualizuj do {{version}}',
        loading: 'Ładowanie...',
        okBtn: 'Tak',
        cancelBtn: 'Nie',
        restart: 'Uruchomić ponownie {{name}}?',
        stop: 'Zatrzymać {{name}}?',
        stopDesc:
          'Usługa zatrzyma się teraz. Uruchamianie przy starcie to osobny przełącznik i pozostaje bez zmian.',
        update: 'Zaktualizować {{name}} do {{version}}?',
        updateDesc: 'Usługa uruchomi się ponownie, jeśli działa. Logowanie zostanie zachowane.',
        notInstall: '{{name}} nie jest zainstalowany.',
        install: 'Zainstaluj',
        installing: 'Instalowanie',
        installFailed: 'Instalacja nie powiodła się',
        retry: 'Spróbuj ponownie',
        notRunning: '{{name}} nie działa. Uruchom go, aby kontynuować.',
        run: 'Uruchom',
        boot: 'Uruchamiaj przy starcie',
        bootDesc: 'Uruchamiaj {{name}} przy starcie KVM.',
        control: 'Serwer sterujący',
        connected: 'Połączono',
        disconnected: 'Nie połączono',
        deviceName: 'Nazwa urządzenia',
        deviceIP: 'IP urządzenia',
        account: 'Konto',
        version: 'Wersja',
        uptime: 'Czas działania',
        peers: 'Węzły',
        noPeers: 'Brak węzłów.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Pamięć',
        daemonRss: 'Usługa',
        group: 'Grupa dodatków',
        high: 'spowalniana powyżej {{size}}',
        max: 'zatrzymywana przez jądro powyżej {{size}}',
        noGroup: 'Brak grupy pamięci dodatków na tej płytce.',
        uninstall: 'Odinstaluj {{name}}',
        uninstallDesc:
          'Czy na pewno chcesz odinstalować {{name}}? Dane logowania pozostaną na płytce.',
        blocked:
          '{{other}} działa lub uruchamia się przy starcie. Naraz może działać tylko jeden VPN: najpierw zatrzymaj {{other}} i wyłącz jego uruchamianie przy starcie.',
        swap: {
          title: 'Plik wymiany',
          tip: 'Jeśli usłudze brakuje pamięci, spróbuj włączyć plik wymiany. Ustawia się go w „Ustawienia > Wydajność”.'
        },
        copy: 'Kopiuj',
        copied: 'Skopiowano link',
        copyFailed: 'Nie udało się skopiować linku. Zaznacz go i skopiuj ręcznie.',
        open: 'Otwórz',
        checkAgain: 'Sprawdź ponownie',
        notSignedIn: 'Jeszcze nie zalogowano. Dokończ logowanie przez link i sprawdź ponownie.',
        checkFailed: 'Nie udało się sprawdzić stanu logowania',
        loginWaiting: 'Strona sprawdza co kilka sekund i przejdzie dalej po zalogowaniu.',
        uninstallFailed: 'Odinstalowanie nie powiodło się',
        loginFailed: 'Logowanie nie powiodło się'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Pobierz',
        package: 'pakiet instalacyjny',
        unzip: 'i wypakuj pliki',
        notLogin:
          'Urządzenie nie zostało jeszcze powiązane. Zaloguj się i powiąż to urządzenie ze swoim kontem.',
        urlPeriod: 'Ten URL jest ważny przez 10 minut',
        login: 'Zaloguj',
        logout: 'Wyloguj',
        logoutDesc: 'Czy na pewno chcesz się wylogować?',
        manualIntro: 'Albo zainstaluj ręcznie przez SSH:',
        copyBinaries: 'Skopiuj tailscale i tailscaled do {{dir}} na IronKVM',
        linksFile: 'W tym samym katalogu utwórz plik o nazwie links z tymi dwoma wierszami:',
        rebootRefresh: 'Uruchom ponownie IronKVM, a potem odśwież tę stronę'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'To urządzenie nie dołączyło jeszcze do sieci NetBird. Dołącz za pomocą klucza konfiguracyjnego lub zaloguj się przez SSO.',
        setupKey: 'Klucz konfiguracyjny',
        setupKeyPlaceholder: 'Wklej klucz konfiguracyjny z panelu NetBird',
        join: 'Dołącz',
        or: 'lub',
        sso: 'Zaloguj przez SSO',
        urlPeriod: 'Ten URL jest ważny przez 10 minut',
        logout: 'Wyrejestruj',
        logoutDesc:
          'Wyrejestrowanie usuwa ten węzeł z konta NetBird i kasuje jego konfigurację na urządzeniu. Ponowne dołączenie wymaga klucza konfiguracyjnego lub logowania SSO, a węzeł może dostać nowy adres IP. Kontynuować?',
        joinFailed: 'Nie udało się dołączyć do sieci'
      },
      update: {
        title: 'Sprawdź aktualizacje',
        queryFailed: 'Uzyskanie wersji nie powiodło się',
        updateFailed: 'Aktualizacja nie powiodła się. Spróbuj ponownie.',
        isLatest: 'Oprogramowanie jest aktualne.',
        available: 'Aktualizacja jest dostępna. Czy na pewno chcesz dokonać aktualizacji?',
        updating: 'Aktualizacja rozpoczęta. Proszę czekać...',
        confirm: 'Potwierdź',
        cancel: 'Anuluj',
        preview: 'Podgląd aktualizacji',
        previewDesc: 'Uzyskaj wcześniejszy dostęp do nowych funkcji i ulepszeń',
        previewTip:
          'Należy pamiętać, że wersje poglądowe mogą zawierać błędy lub niekompletną funkcjonalność!',
        customServer: {
          title: 'Niestandardowy serwer aktualizacji',
          desc: 'Sprawdzaj dostępność aktualizacji online i pobieraj je ze wskazanego serwera',
          invalidUrl:
            'Wprowadź prawidłowy adres katalogu serwera HTTP lub HTTPS, bez zapytania, fragmentu ani pliku latest.json.',
          loadFailed: 'Nie udało się wczytać konfiguracji serwera aktualizacji.',
          saveFailed: 'Nie udało się zapisać konfiguracji serwera aktualizacji.',
          saved: 'Konfiguracja serwera aktualizacji została zapisana.',
          save: 'Zapisz',
          confirmTitle: 'Użyć niestandardowego serwera aktualizacji?',
          confirmDesc:
            'SHA-512 sprawdza jedynie, czy pakiet jest zgodny z manifestem dostarczonym przez ten serwer. Nie potwierdza, że pakiet jest oficjalnym wydaniem IronKVM. Wadliwy lub złośliwy serwer może unieruchomić urządzenie, spowodować utratę danych lub naruszyć bezpieczeństwo systemu.',
          confirm: 'Użyj mimo to',
          useSipeed: 'Użyj oficjalnego serwera Sipeed',
          previewDisabled:
            'Aktualizacje w wersji testowej są niedostępne, gdy włączony jest niestandardowy serwer aktualizacji.'
        },
        offline: {
          chooseFile: 'Wybierz plik',
          installing: 'Przesłano. Instalowanie...',
          noFile: 'Nie wybrano pliku',
          title: 'Aktualizacje offline',
          desc: 'Aktualizacja poprzez lokalny pakiet instalacyjny',
          upload: 'Prześlij',
          checksumPlaceholder: 'Suma kontrolna SHA-256 (opcjonalnie)',
          invalidChecksum: 'Suma kontrolna SHA-256 musi zawierać 64 znaki szesnastkowe.',
          checksumMismatch: 'Weryfikacja SHA-256 nie powiodła się. Pakiet może być uszkodzony.',
          invalidName: 'Nieprawidłowy format nazwy pliku. Proszę pobrać z wydań GitHub.',
          updateFailed: 'Aktualizacja nie powiodła się. Spróbuj ponownie.'
        },
        updateTo: 'Aktualizuj do {{version}}',
        updateConfirmDesc:
          'Urządzenie zainstaluje aktualizację i uruchomi ponownie swój serwer. Strona przeładuje się, gdy serwer wróci.',
        releaseNotes: 'Informacje o wydaniu'
      },
      account: {
        title: 'Konto',
        webAccount: 'Nazwa konta web',
        role: 'Rola',
        roles: { admin: 'Administrator', user: 'Użytkownik' },
        password: 'Hasło',
        updateBtn: 'Update',
        logoutBtn: 'Wyloguj',
        logoutDesc: 'Czy na pewno chcesz się wylogować?',
        okBtn: 'Tak',
        cancelBtn: 'Nie',
        users: {
          title: 'Użytkownicy',
          create: 'Utwórz użytkownika',
          enabled: 'Włączony',
          disabled: 'Wyłączony',
          deviceOwner: 'Właściciel urządzenia',
          resetPassword: 'Resetuj hasło',
          delete: 'Usuń',
          deleteConfirm: 'Usunąć tego użytkownika i unieważnić wszystkie jego sesje?',
          created: 'Utworzono użytkownika',
          deleted: 'Usunięto użytkownika',
          passwordUpdated: 'Zaktualizowano hasło',
          loadFailed: 'Nie udało się wczytać użytkowników',
          saveFailed: 'Nie udało się zapisać użytkownika',
          deleteFailed: 'Nie udało się usunąć użytkownika'
        }
      },
      apiKeys: {
        mcpNote: 'Te klucze nie działają dla MCP, który ma własny klucz na stronie MCP.',
        metricsUrl: 'URL metryk',
        monitoring: 'Monitorowanie',
        monitoringDesc:
          'Prometheus odczytuje metryki kluczem API z tej strony, wysyłanym jako token Bearer. Może je czytać każda rola.',
        scrapeConfig: 'Konfiguracja scrape dla Prometheus',
        title: 'Klucze API',
        description:
          'Klucz działa w imieniu swojego właściciela, z jego rolą. Wysyłaj go jako Authorization: Bearer <key> dla metryk i API lub jako X-Auth-Token dla Redfish.',
        name: 'Nazwa',
        namePlaceholder: 'Do czego służy klucz, np. prometheus',
        nameRequired: 'Nadaj kluczowi nazwę',
        nameTooLong: 'Nazwa może mieć najwyżej 64 znaki',
        unnamed: '(bez nazwy)',
        create: 'Utwórz klucz',
        created: 'Utworzono',
        owner: 'Właściciel',
        empty: 'Brak kluczy API',
        newKeyTitle: 'Twój nowy klucz API',
        newKeyWarning:
          'Skopiuj klucz teraz. Nie jest przechowywany i nie można go ponownie wyświetlić. Jeśli go zgubisz, unieważnij go i utwórz nowy.',
        copy: 'Kopiuj',
        copied: 'Skopiowano',
        copyFailed: 'Kopiowanie nie powiodło się. Skopiuj ręcznie.',
        done: 'Gotowe',
        revoke: 'Unieważnij',
        revokeConfirmTitle: 'Unieważnić ten klucz API?',
        revokeConfirmDesc: 'Wszystko, co używa „{{name}}”, natychmiast przestanie działać.',
        revoked: 'Unieważniono klucz API',
        loadFailed: 'Nie udało się wczytać kluczy API',
        createFailed: 'Nie udało się utworzyć klucza API',
        revokeFailed: 'Nie udało się unieważnić klucza API',
        cancelBtn: 'Anuluj'
      }
    },
    picoclaw: {
      title: 'PicoClaw Asystent',
      empty: 'Otwórz panel i rozpocznij zadanie.',
      inputPlaceholder: 'Opisz, co chcesz, aby PicoClaw zrobił',
      newConversation: 'Nowa rozmowa',
      processing: 'Przetwarzanie...',
      agent: {
        defaultTitle: 'Asystent ogólny',
        defaultDescription: 'Ogólna pomoc dotycząca czatu, wyszukiwania i przestrzeni roboczej.',
        kvmTitle: 'Zdalne sterowanie',
        kvmDescription: 'Sterowanie zdalnym hostem poprzez IronKVM.',
        switched: 'Rola agenta została zmieniona',
        switchFailed: 'Nie udało się zmienić roli agenta'
      },
      send: 'Wyślij',
      cancel: 'Anuluj',
      status: {
        connecting: 'Łączenie z bramką...',
        connected: 'Sesja PicoClaw połączona',
        disconnected: 'Sesja PicoClaw zamknięta',
        stopped: 'Wysłano żądanie zatrzymania',
        runtimeStarted: 'Runtime PicoClaw uruchomiony',
        runtimeStartFailed: 'Nie udało się uruchomić runtime PicoClaw',
        runtimeStopped: 'Runtime PicoClaw zatrzymany',
        runtimeStopFailed: 'Nie udało się zatrzymać runtime PicoClaw',
        controlSwitchedToMCP: 'Sterowanie przełączono na zewnętrzną usługę MCP'
      },
      connection: {
        runtime: {
          checking: 'Sprawdzam',
          restoring: 'Przywracanie PicoClaw',
          ready: 'Runtime gotowy',
          stopped: 'Runtime zatrzymany',
          blockedByMCP: 'Zewnętrzne sterowanie MCP jest aktywne',
          readyBlockedByMCP:
            'Runtime działa, ale wejściem urządzenia steruje teraz zewnętrzny MCP.',
          readyWithoutControl:
            'Runtime działa. Przed ponownym połączeniem przekaż PicoClaw sterowanie urządzeniem.',
          unavailable: 'Runtime niedostępny',
          configError: 'Błąd konfiguracji'
        },
        transport: {
          connecting: 'Łączenie',
          connected: 'Połączono',
          disconnected: 'Rozłączono',
          reconnect: 'Połącz ponownie',
          reconnectDescription: 'Połącz ponownie z działającą sesją PicoClaw.',
          reconnectBlocked: 'PicoClaw potrzebuje sterowania urządzeniem, aby połączyć się ponownie.'
        },
        run: {
          idle: 'Bezczynność',
          busy: 'Zajęty'
        }
      },
      message: {
        toolAction: 'Akcja',
        observation: 'Obserwacja',
        screenshot: 'Zrzut ekranu'
      },
      overlay: {
        locked: 'PicoClaw steruje urządzeniem. Wprowadzanie ręczne zostało wstrzymane.'
      },
      control: {
        picoclaw: 'Sterowanie urządzeniem: PicoClaw',
        picoclawDescription:
          'PicoClaw może wysyłać dane z klawiatury i myszy. Ręczne wprowadzanie może zostać wstrzymane.',
        mcp: 'Sterowanie urządzeniem: zewnętrzny MCP',
        mcpDescription:
          'Zewnętrzny MCP może zapisywać do urządzenia. PicoClaw nie przejmie wprowadzania.',
        off: 'Sterowanie urządzeniem: wyłączone',
        offDescription:
          'AI nie będzie wysyłać danych z klawiatury ani myszy. Ręczne sterowanie pozostaje dostępne.',
        transitioning: 'Sterowanie urządzeniem: przełączanie',
        transitioningDescription: 'Sterowanie urządzeniem jest synchronizowane. Proszę czekać.',
        grant: 'Przekaż sterowanie',
        release: 'Zwolnij',
        releasing: 'Zwalnianie...',
        switching: 'Przełączanie...',
        releasingLabel: 'Sterowanie urządzeniem: zwalnianie',
        releasingDescription:
          'Sterowanie urządzeniem jest oddawane. PicoClaw zatrzymał bieżące zapisy.',
        granted: 'Sterowanie PicoClaw przyznane',
        released: 'Sterowanie PicoClaw zwolnione',
        grantFailed: 'Nie udało się przyznać sterowania PicoClaw',
        releaseFailed: 'Nie udało się zwolnić sterowania PicoClaw',
        grantConfirmTitle: 'Przełączyć sterowanie urządzeniem na PicoClaw?',
        grantConfirmDesc: 'Zapisy urządzenia przez zewnętrzny MCP zostaną przerwane.'
      },
      install: {
        install: 'Zainstaluj PicoClaw',
        installing: 'Instalowanie PicoClaw',
        success: 'PicoClaw zainstalowano pomyślnie',
        failed: 'Nie udało się zainstalować PicoClaw',
        uninstalling: 'Odinstalowywanie runtime...',
        uninstalled: 'Runtime został pomyślnie odinstalowany.',
        uninstallFailed: 'Odinstalowanie nie powiodło się.',
        requiredTitle: 'PicoClaw nie jest zainstalowany',
        requiredDescription: 'Zainstaluj PicoClaw przed uruchomieniem runtime PicoClaw.',
        progressDescription: 'PicoClaw jest pobierany i instalowany.',
        stages: {
          preparing: 'Przygotowanie',
          downloading: 'Pobieranie',
          extracting: 'Wypakowywanie',
          verifying: 'Weryfikowanie',
          installing: 'Instalowanie',
          installed: 'Zainstalowano',
          install_timeout: 'Upłynął limit czasu',
          install_failed: 'Niepowodzenie'
        }
      },
      model: {
        requiredTitle: 'Wymagana jest konfiguracja modelu',
        requiredDescription: 'Skonfiguruj model PicoClaw przed użyciem czatu PicoClaw.',
        docsTitle: 'Przewodnik konfiguracji',
        docsDesc: 'Obsługiwane modele i protokoły',
        menuLabel: 'Skonfiguruj model',
        modelIdentifier: 'Identyfikator modelu',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Klucz API',
        apiKeyPlaceholder: 'Wprowadź klucz API modelu',
        save: 'Zapisz',
        saving: 'Zapisywanie',
        saved: 'Konfiguracja modelu została zapisana',
        saveFailed: 'Nie udało się zapisać konfiguracji modelu',
        invalid: 'Identyfikator modelu, API Base URL i klucz API są wymagane'
      },
      uninstall: {
        menuLabel: 'Odinstaluj',
        confirmTitle: 'Odinstaluj PicoClaw',
        confirmContent:
          'Czy na pewno chcesz odinstalować PicoClaw? Spowoduje to usunięcie pliku wykonywalnego i wszystkich plików konfiguracyjnych.',
        confirmOk: 'Odinstaluj',
        confirmCancel: 'Anuluj'
      },
      history: {
        title: 'Historia',
        loading: 'Ładowanie sesji...',
        emptyTitle: 'Nie ma jeszcze historii',
        emptyDescription: 'Tutaj pojawią się poprzednie sesje PicoClaw.',
        loadFailed: 'Nie udało się załadować historii sesji',
        deleteFailed: 'Nie udało się usunąć sesji',
        deleteConfirmTitle: 'Usuń sesję',
        deleteConfirmContent: 'Czy na pewno chcesz usunąć „{{title}}”?',
        deleteConfirmOk: 'Usuń',
        deleteConfirmCancel: 'Anuluj',
        messageCount_one: '{{count}} wiadomość',
        messageCount_few: '{{count}} wiadomości',
        messageCount_many: '{{count}} wiadomości',
        messageCount_other: '{{count}} wiadomości',
        messageCount: '{{count}} wiadomości'
      },
      config: {
        startRuntime: 'Uruchom PicoClaw',
        stopRuntime: 'Zatrzymaj PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Przełączyć sterowanie na PicoClaw?',
        enableConfirmDesc: 'Uruchomienie PicoClaw wyłączy zewnętrzną usługę MCP.',
        enableConfirmOk: 'Uruchom PicoClaw',
        enableConfirmCancel: 'Anuluj',
        title: 'Uruchom PicoClaw',
        description: 'Uruchom runtime, aby rozpocząć korzystanie z asystenta PicoClaw.',
        switchFromMCP: 'Przełącz na PicoClaw i uruchom',
        takeoverAndStart: 'Przejmij i uruchom'
      }
    },
    error: {
      title: 'Wystąpił problem',
      refresh: 'Odśwież',
      panel: 'Ta część strony przestała działać',
      retry: 'Ponów'
    },
    fullscreen: {
      toggle: 'Przełącz tryb pełnoekranowy'
    },
    input: {
      disconnected: 'Klawiatura i mysz nie są połączone',
      disconnectedTls:
        'Przeglądarka odrzuciła bezpieczne połączenie, które przenosi klawiaturę i mysz, i zrobiła to bez pytania. Certyfikat wygenerowany przez to urządzenie nie jest jeszcze zaufany. Otwórz ten adres w nowej karcie, zaakceptuj certyfikat, a potem odśwież stronę. Niezawodnym rozwiązaniem jest zainstalowanie certyfikatu.',
      disconnectedNever:
        'Nie udało się otworzyć połączenia, które przenosi klawiaturę i mysz. Reszta strony działa, bo z niego nie korzysta. Sprawdź, czy nic między Tobą a urządzeniem go nie blokuje.',
      disconnectedDropped:
        'Połączenie, które przenosi klawiaturę i mysz, zostało utracone i nie wróciło. Po restarcie łączy się samo; jeśli to nie mija, odśwież stronę.',
      hidDisabled: 'HID jest wyłączone na tym urządzeniu (/boot/disable_hid).',
      keyFailed: 'Nie udało się wysłać klawisza.'
    },
    speaker: { title: 'Głośnik', unmute: 'Włącz dźwięk', mute: 'Wycisz' },
    upstream: {
      check: 'Sprawdź aktualizacje',
      updateTo: 'Zaktualizuj do {{version}}',
      confirm: 'Zaktualizować {{name}} do {{version}}?',
      confirmDesc:
        'Nowe wydanie jest pobierane z GitHuba i sprawdzane z opublikowanymi sumami kontrolnymi. Jeśli coś się nie powiedzie, zostaje obecna wersja.',
      ok: 'Zaktualizuj',
      upToDate: 'Aktualne',
      builtIn: 'wbudowana',
      checkFailed: 'Nie udało się sprawdzić aktualizacji: {{error}}',
      unverifiable: 'Wersja {{version}} nie jest oferowana: {{reason}}',
      inUse: 'Nie można teraz zaktualizować: {{reason}}',
      running: 'Aktualizacja do {{version}}...',
      done: 'Zaktualizowano {{name}} do {{version}}',
      failed: 'Ostatnia aktualizacja nie powiodła się: {{error}}'
    },
    menu: {
      collapse: 'Zwiń menu',
      expand: 'Rozwiń Menu',
      more: 'Więcej',
      media: 'Nośniki',
      tools: 'Narzędzia',
      text: 'Tekst',
      advanced: 'Zaawansowane',
      mediaMounted: 'Zamontowane',
      mediaLibrary: 'Biblioteka',
      mediaBoot: 'Rozruch',
      textToHost: 'Do hosta',
      textFromHost: 'Z hosta'
    },
    ion: {
      checking: 'Sprawdzanie pamięci wideo przed uruchomieniem strumienia...',
      warn: 'Mało pamięci wideo. Jeden restart serwera ją wyczerpie. Uruchom ponownie w dogodnej chwili.',
      criticalTitle: 'Za mało pamięci wideo, aby uruchomić strumień',
      criticalBody:
        'Uruchomienie wideo wyczerpie zarezerwowaną pamięć i zatrzyma serwer. Wszystkie inne funkcje nadal działają, w tym sterowanie zasilaniem i ponowne uruchamianie. Tę pamięć odzyskuje tylko ponowne uruchomienie IronKVM.',
      criticalContinue: 'Uruchom wideo mimo to',
      criticalReboot: 'Uruchom ponownie IronKVM',
      criticalRebooting: 'Ponowne uruchamianie...'
    }
  }
};

export default pl;
