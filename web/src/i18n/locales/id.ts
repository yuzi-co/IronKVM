const id = {
  translation: {
    feedback: {
      enabled: '{{name}} diaktifkan',
      disabled: '{{name}} dinonaktifkan',
      failed: 'Permintaan gagal. Coba lagi.',
      network: 'Perangkat tidak dapat dihubungi. Periksa koneksi dan coba lagi.',
      saved: 'Tersimpan',
      timeout: 'Perangkat terlalu lama menjawab. Coba lagi.'
    },
    common: {
      copy: 'Salin',
      copied: 'Tersalin',
      copyFailed: 'Gagal menyalin. Pilih teks dan salin secara manual.',
      notUpdating: 'Tidak diperbarui: penyegaran terakhir gagal.',
      off: 'Mati',
      running: 'Berjalan',
      save: 'Simpan',
      cancel: 'Batal',
      delete: 'Hapus',
      remove: 'Buang'
    },
    head: {
      desktop: 'Desktop jarak jauh',
      login: 'Masuk',
      changePassword: 'Ubah Sandi',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      passwordChanged: 'Kata sandi diubah. Masuk dengan kata sandi baru.',
      cookieRejected:
        'Browser menolak menyimpan sesi. Cookie yang tertinggal dari sesi HTTPS sebelumnya tidak dapat diganti melalui http biasa. Hapus cookie untuk alamat ini, atau buka jendela pribadi, lalu masuk kembali.',
      login: 'Masuk',
      placeholderUsername: 'Silahkan masukkan username',
      placeholderPassword: 'Silahkan masukkan password',
      placeholderCurrentPassword: 'Kata sandi saat ini',
      placeholderPassword2: 'Silahkan masukkan password again',
      noEmptyUsername: 'nama user tidak boleh kosong',
      noEmptyPassword: 'sandi  tidak boleh kosong',
      passwordLength: 'Kata sandi harus terdiri dari 8 hingga 72 karakter',
      noAccount:
        'Gagal mendapatkan informasi user, silahkan segarkan halaman atau atur ulang sandi',
      invalidUser: 'invalid username or password',
      locked: 'Terlalu banyak login, silakan coba lagi nanti',
      globalLocked: 'Sistem dalam perlindungan, silakan coba lagi nanti',
      error: 'terjadi kesalahan tak terduga',
      invalidCurrentPassword: 'Kata sandi saat ini salah',
      changePassword: 'Ganti Sandi',
      changePasswordDesc: 'Untuk keamanan perangkat Anda, silakan ubah kata sandi masuk web.',
      differentPassword: 'sandi tidak sesuai',
      illegalUsername: 'ada karakter ilegal pada nama user',
      illegalPassword: 'ada karakter ilegal pada sandi',
      forgetPassword: 'Lupa Sandi',
      ok: 'Ok',
      cancel: 'Batalkan',
      loginButtonText: 'Masuk',
      tips: {
        reset1:
          'To reset the passwords, pressing and holding the BOOT button on the IronKVM for 10 seconds.',
        reset3: 'Akun web default:',
        reset4: 'Akun SSH default:',
        change1: 'Perhatikan bahwa tindakan ini akan mengubah kata sandi berikut:',
        change2: 'Kata sandi login web',
        change3: 'Kata sandi root sistem (kata sandi login SSH)',
        change4: 'Untuk mengatur ulang kata sandi, tekan dan tahan tombol BOOT pada IronKVM.',
        resetDocs: 'Untuk langkah terperinci, lihat dokumentasi perangkat keras:',
        hardwareDocs: 'Wiki Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Konfigurasi Wi-Fi untuk IronKVM',
      success: 'Please check the network status of IronKVM and visit the new IP address.',
      failed: 'Operasi gagal, silakan coba lagi.',
      invalidMode:
        'Mode saat ini tidak mendukung pengaturan jaringan. Silakan buka perangkat Anda dan aktifkan mode konfigurasi Wi-Fi.',
      confirmBtn: 'Ok',
      finishBtn: 'Selesai',
      ap: {
        authTitle: 'Otentikasi Diperlukan',
        authDescription: 'Silakan masukkan kata sandi AP untuk melanjutkan',
        authFailed: 'Kata sandi AP tidak valid',
        passPlaceholder: 'AP kata sandi',
        verifyBtn: 'Verifikasi'
      },
      ssidRequired: 'Masukkan nama jaringan, maksimal 32 karakter',
      passwordLength: 'Kata sandi 8 sampai 63 karakter. Biarkan kosong untuk jaringan terbuka.',
      passwordOptional: 'Kata sandi (kosong untuk jaringan terbuka)',
      lost: 'Papan berhenti merespons. Mungkin sudah bergabung ke jaringan dan menutup hotspot penyiapannya. Jika hotspot muncul lagi, penggabungan gagal: sambungkan lagi dan coba ulang.',
      done: 'Penyiapan selesai. Sambungkan kembali perangkat ini ke jaringan biasa Anda dan buka papan di alamat barunya.'
    },
    screen: {
      viewOnly: 'Hanya lihat',
      viewOnlyTip:
        'Tab ini berhenti mengirim keyboard dan mouse ke host. Skrip, jiggler mouse, dan penonton lain tidak terpengaruh.',
      viewOnlyOff: 'Matikan hanya lihat',
      viewOnlyBlocked: 'Hanya lihat aktif, jadi tidak ada yang dikirim ke host',
      pauseHidden: 'Jeda saat tab tersembunyi',
      pauseHiddenTip:
        'Menghentikan video dan suaranya beberapa detik setelah tab ini disembunyikan, dan memulainya lagi saat Anda kembali.',
      screenshot: 'Tangkapan layar',
      screenshotTip: 'Menyimpan layar host sebagai PNG dalam ukuran tangkapan penuh.',
      screenshotFailed: 'Tangkapan layar gagal',
      stream: {
        ok: 'gambar OK',
        noSignal: 'tidak ada sinyal',
        failed: 'stream gagal'
      },
      codecNoWebrtcHevc: 'Browser ini tidak dapat menerima H.265 melalui WebRTC',
      codecNoHevc: 'Browser ini tidak dapat mendekode H.265',
      codecNote:
        'Papan hanya punya satu encoder, jadi ini mengubah stream untuk semua penonton. Sambungkan ulang agar berlaku pada sesi WebRTC yang berjalan.',
      codec: 'Codec',
      updateFailed: 'Pengaturan tidak diterapkan',
      scale: 'Skala',
      title: 'Layar',
      video: 'Mode Video',
      videoDirectTips: 'Aktifkan HTTPS di "Pengaturan > Perangkat" untuk menggunakan mode ini',
      resolution: 'Resolusi',
      ocr: {
        title: 'Baca Teks (OCR)',
        tips: 'Teks dikenali di browser ini. Anda dapat memperbaikinya sebelum menyalinnya.',
        hint: 'Seret di atas teks yang ingin dibaca. Tekan Esc untuk membatalkan.',
        noPicture: 'Tunggu video, lalu seret di atas teks yang ingin dibaca.',
        cancel: 'Batal',
        language: 'Bahasa',
        languages: {
          eng: 'Inggris'
        },
        preview: 'Area yang dipilih',
        capturing: 'Menangkap layar...',
        loading: 'Memuat pengenalan teks...',
        recognizing: 'Membaca teks...',
        noText: 'Tidak ada teks yang ditemukan di area yang dipilih.',
        copy: 'Salin',
        copied: 'Disalin ke clipboard',
        copyFailed: 'Gagal menyalin ke clipboard',
        selectAgain: 'Pilih Lagi',
        unsupported:
          'Browser ini tidak dapat menjalankan pengenalan teks. Fitur ini memerlukan WebAssembly SIMD, yang dimiliki browser saat ini.',
        captureFailed: 'Gagal menangkap layar.',
        outside: 'Area yang dipilih berada di luar gambar.',
        recognizeFailed: 'Pengenalan teks gagal.'
      },
      controlRegion: {
        title: 'Kalibrasi Tetikus',
        description:
          'Gunakan pengaturan ini saat perangkat yang dikontrol menggunakan resolusi selain 16:9 dan posisi kursor tidak sejajar secara horizontal atau vertikal.',
        off: 'Nonaktif',
        auto: 'Otomatis',
        autoWarning:
          'Kalibrasi mungkin gagal jika aplikasi pengguna menggunakan latar belakang hitam pekat.',
        manual: 'Manual',
        selectedResolution: 'Resolusi Area Terpilih',
        unused: 'Tidak digunakan',
        originalResolution: 'Resolusi Asli',
        selectResolution: 'Pilih resolusi asli',
        addResolution: 'Tambahkan resolusi khusus',
        add: 'Tambah',
        duplicateResolution: 'Resolusi ini sudah ada.',
        width: 'Lebar',
        height: 'Tinggi',
        apply: 'Hitung dan Terapkan',
        invalidResolution: 'Masukkan resolusi asli yang valid setelah video siap.',
        select: 'Pilih Area',
        clear: 'Pulihkan Otomatis',
        saveFailed: 'Gagal menyimpan area input.',
        tooSmall: 'Area yang dipilih terlalu kecil.',
        previewUnavailable: 'Pratinjau tidak tersedia',
        clearConfirm: 'Pulihkan deteksi batas hitam otomatis?',
        dragHint: 'Seret untuk memilih area desktop jarak jauh',
        finish: 'Selesai',
        confirm: 'Konfirmasi',
        cancel: 'Batal'
      },
      auto: 'Otomatis',
      autoTips:
        'Tearing layar atau offset tetikus dapat terjadi pada resolusi tertentu. Pertimbangkan untuk menyesuaikan resolusi host jarak jauh atau menonaktifkan mode otomatis.',
      fps: 'FPS',
      customizeFps: 'Sesuaikan',
      quality: 'Kualitas',
      qualityLossless: 'Terbaik',
      qualityHigh: 'Tinggi',
      qualityMedium: 'Sedang',
      qualityLow: 'Rendah',
      frameDetect: 'Deteksi bingkai',
      frameDetectTip:
        'Hitung selisih antar frame. Hentikan transmisi aliran video saat tidak ada perubahan yang terdeteksi di layar host jarak jauh.',
      resetHdmi: 'Atur ulang HDMI',
      mixedH264: {
        title: 'Konflik aliran H.264',
        description:
          'H.264 Direct dan H.264 WebRTC sedang digunakan secara bersamaan. Hal ini dapat menyebabkan layar robek atau video rusak. Harap gunakan hanya satu mode H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Koneksi WebRTC gagal',
        description: 'Periksa koneksi jaringan atau ganti mode video.'
      },
      captureStatus: {
        hdmiError: 'Kesalahan layar HDMI',
        unsupportedResolution: 'Resolusi saat ini tidak didukung',
        retrieving: 'Mengambil layar...',
        changingResolution: 'Mengganti resolusi...',
        updateFailed: 'Layar tidak dapat diperbarui saat ini',
        videoError: 'Kesalahan tampilan video',
        noHdmi: 'Sinyal HDMI tidak terdeteksi',
        unavailable: 'Layar tidak dapat ditampilkan saat ini'
      },
      directConnectionFailed: 'Koneksi aliran video gagal'
    },
    keyboard: {
      close: 'Tutup',
      title: 'Keyboard',
      paste: 'Tempel',
      tips: 'Mengetik teks di host sebagai penekanan tombol. Pilih tata letak keyboard yang dipakai host.',
      placeholder: 'Silahkan isi',
      submit: 'Kirimkan',
      virtual: 'Keyboard',
      readClipboard: 'Membaca dari Papan Klip',
      clipboardPermissionDenied:
        'Izin papan klip ditolak. Harap izinkan akses clipboard di browser Anda.',
      clipboardReadError: 'Gagal membaca papan klip',
      mediaKeys: {
        title: 'Tombol media',
        mute: 'Bisukan',
        volumeDown: 'Kecilkan volume',
        volumeUp: 'Besarkan volume',
        previous: 'Lagu sebelumnya',
        playPause: 'Putar atau jeda',
        next: 'Lagu berikutnya',
        stop: 'Hentikan'
      },
      pasting: {
        layout: 'Tata letak keyboard di host',
        layouts: {
          us: 'Inggris (AS)',
          uk: 'Inggris (Britania Raya)',
          de: 'Jerman',
          fr: 'Prancis',
          es: 'Spanyol',
          it: 'Italia',
          ptBr: 'Portugis (Brasil)',
          se: 'Swedia / Finlandia',
          ru: 'Rusia',
          ja: 'Jepang',
          ko: 'Korea'
        },
        speed: 'Kecepatan mengetik',
        speeds: {
          fast: 'Cepat',
          normal: 'Normal',
          slow: 'Lambat'
        },
        estimate: 'Waktu mengetik: sekitar {{duration}}',
        untypeable: 'Karakter yang tidak bisa diketik dengan tata letak ini: {{count}}',
        untypeableAt: 'baris {{line}}, kolom {{column}}',
        skipUntypeable: 'Ketik sisanya',
        shortcut: '{{shortcut}} langsung mengetik isi clipboard di host.',
        clipboardUnavailable:
          'Browser hanya mengizinkan halaman membaca clipboard melalui HTTPS. Tempel teks ke kotak dengan Ctrl+V.',
        clipboardEmpty: 'Clipboard tidak berisi teks.',
        tooLong: 'Teks terlalu panjang. Batasnya {{max}} karakter.',
        inProgress: 'Tempelan lain sedang diketik.',
        typing: 'Mengetik di host',
        done: 'Teks selesai diketik',
        canceled: 'Tempel dibatalkan',
        failed: 'Tempel gagal',
        cancel: 'Batal',
        controlBusy: 'Pengontrol lain sedang memakai keyboard.',
        hidError: 'Penekanan tombol tidak dapat dikirim ke host.'
      },
      shortcut: {
        sendFailed: 'Tidak terkirim: koneksi input terputus',
        title: 'Pintasan',
        custom: 'Adat',
        capture: 'Klik di sini untuk mengambil pintasan',
        clear: 'Jelas',
        save: 'Simpan',
        captureTips:
          'Menangkap tombol tingkat sistem (seperti tombol Windows) memerlukan izin layar penuh.',
        enterFullScreen: 'Beralih ke mode layar penuh.'
      },
      leaderKey: {
        saveFailed: 'Gagal menyimpan tombol leader',
        title: 'Tombol Leader',
        desc: 'Lewati batasan browser dan kirim pintasan sistem langsung ke host jarak jauh.',
        howToUse: 'Cara Menggunakan',
        simultaneous: {
          title: 'Mode Simultan',
          desc1: 'Tekan dan tahan tombol Leader, lalu tekan pintasan.',
          desc2: 'Intuitif, tetapi mungkin bertentangan dengan pintasan sistem.'
        },
        sequential: {
          title: 'Mode Berurutan',
          desc1:
            'Tekan tombol Leader → tekan pintasan secara berurutan → tekan tombol Leader lagi.',
          desc2: 'Memerlukan lebih banyak langkah, namun sepenuhnya menghindari konflik sistem.'
        },
        enable: 'Aktifkan tombol Leader',
        tip: 'Saat ditetapkan sebagai tombol Leader, tombol ini hanya berfungsi sebagai pemicu pintasan dan kehilangan perilaku defaultnya.',
        placeholder: 'Tekan tombol Leader',
        shiftRight: 'Shift kanan',
        ctrlRight: 'Ctrl kanan',
        metaRight: 'Win kanan',
        submit: 'Kirimkan',
        recorder: {
          rec: 'REKAM',
          activate: 'Aktifkan tombol',
          input: 'Silakan tekan pintasan...'
        }
      }
    },
    mouse: {
      jiggler: 'Penggerak mouse',
      jigglerMouse: 'Mouse',
      jigglerF15: 'Tombol F15',
      jigglerShift: 'Tombol Shift',
      jigglerCtrl: 'Tombol Ctrl',
      jigglerF15Tip: 'F15 paling tidak mengganggu: tidak dipakai sistem atau aplikasi umum mana pun',
      title: 'Tikus',
      cursor: 'Gaya kursor',
      default: 'Kursor bawaan',
      pointer: 'Kursor penunjuk',
      cell: 'Kursor cell',
      text: 'Kursor teks',
      grab: 'Kursor ambil',
      hide: 'Sembunyikan kursor',
      mode: 'Mode tetikus',
      absolute: 'Mode absolut',
      relative: 'Mode relatif',
      absoluteShort: 'Absolut',
      relativeShort: 'Relatif',
      touch: 'Mode sentuh',
      touchShort: 'Sentuh',
      absoluteStalled: 'Target mengabaikan tetikus absolut',
      absoluteStalledDesc:
        'Target berhenti menerima laporan tetikus absolut, sehingga gerakan penunjuk hilang. Keyboard tidak terpengaruh. Memulihkan USB biasanya mengatasinya; mode relatif menggunakan endpoint yang berbeda.',
      useRelative: 'Beralih ke mode relatif',
      direction: 'Arah roda gulir',
      scrollUp: 'Sama seperti komputer ini',
      scrollDown: 'Terbalik (gulir alami)',
      speed: 'Kecepatan roda gulir',
      fast: 'Cepat',
      slow: 'Lambat',
      requestPointer:
        'Menggunakan mode relatf. Silakan klik desktop untuk mendapatkan penunjuk tetikus.',
      resetHid: 'Setel ulang HID',
      hidOnly: {
        switchFailed: 'Gagal mengganti mode. Periksa koneksi lalu coba lagi.',
        title: 'Mode hanya HID',
        desc: 'Jika mouse dan keyboard Anda berhenti merespons dan menyetel ulang HID tidak membantu, mungkin ada masalah kompatibilitas antara IronKVM dan perangkat. Coba aktifkan mode HID-Only untuk kompatibilitas yang lebih baik.',
        tip1: 'Mengaktifkan mode HID-Hanya akan melepas U-disk virtual dan jaringan virtual',
        tip2: 'Dalam mode HID-Only, pemasangan gambar dinonaktifkan',
        rebuild: 'Mengganti mode akan membangun ulang koneksi USB. IronKVM tidak dimulai ulang',
        enable: 'Aktifkan mode HID-Hanya',
        disable: 'Nonaktifkan mode HID-Hanya'
      },
      resetHidDone: 'HID USB telah direset',
      resetHidFailed: 'Reset HID USB gagal'
    },
    image: {
      driveLoaded: 'image dimuat',
      driveWarning: 'periksa peringatannya',
      warning: {
        missing:
          'File image telah dihapus. Host tetap membaca salinan lama sampai Anda mengeluarkannya.',
        writable: 'Baca-tulis: host dapat mengubah image ini.',
        tooBigForCd: 'Terlalu besar untuk drive CD ({{size}}, batas {{max}}). Gunakan disk.',
        tooSmallForCd: 'Terlalu kecil untuk drive CD ({{size}}). Gunakan disk.',
        empty: 'File kosong, mungkin karena unggahan atau unduhan yang gagal.'
      },
      delete: 'Hapus',
      inUse: 'Sedang digunakan. Keluarkan sebelum menghapus.',
      retry: 'Coba lagi',
      loadFailed: 'Gagal memuat daftar image',
      readOnlyLocked: 'Keluarkan disk untuk mengubah ini. Berlaku saat image dimasukkan.',
      title: 'Gambar',
      loading: 'Memuat...',
      empty: 'Tidak ada yang ditemukan',
      mountMode: 'Mode pemasangan',
      mountFailed: 'Pemasangan Gagal',
      mountDesc:
        'Di beberapa sistem, perlu mengeluarkan disk virtual pada host jarak jauh sebelum memasang gambar.',
      unmountFailed: 'Pelepasan gagal',
      unmountDesc:
        'Pada beberapa sistem, Anda perlu mengeluarkan secara manual dari host jarak jauh sebelum melepas gambar.',
      refresh: 'Segarkan daftar gambar',
      disk: 'Disk',
      cdrom: 'CD',
      driveEmpty: 'Kosong',
      eject: 'Keluarkan',
      readOnly: 'Hanya baca',
      readOnlyTip: 'Berlaku untuk gambar berikutnya yang dimasukkan ke disk.',
      noDrives: 'Tidak ada drive virtual. Aktifkan disk virtual di Pengaturan.',
      insertFailed: 'Gagal memasukkan',
      ejectFailed: 'Gagal mengeluarkan',
      insertInto: 'Masukkan ke {{drive}}. Klik untuk mengubah.',
      loadedIn: 'Di drive {{drive}}',
      attention: 'Perhatian',
      deleteConfirm: 'Apakah Anda yakin ingin menghapus gambar ini?',
      okBtn: 'Ya',
      cancelBtn: 'Tidak',
      deleteFailed: 'Gagal menghapus',
      ventoy: {
        statusNoKernel: 'Tidak didukung oleh firmware ini',
        statusNotInstalled: 'Belum terpasang',
        statusReady: 'Siap',
        statusSelected: 'Image terpilih: {{count}}',
        statusInDrive: 'Di drive disk, {{size}}',
        noKernel:
          'Kernel firmware ini tidak mendukung device-mapper, jadi Ventoy tidak dapat digunakan sampai image yang mendukungnya dipasang.',
        installDesc: 'Boot host dari beberapa image dalam satu disk, tanpa menyalinnya.',
        install: 'Pasang',
        installing: 'Mengunduh Ventoy, sekitar 20 MB. Ini dapat memakan waktu beberapa menit.',
        needsData: 'Ventoy memerlukan image IronKVM dengan partisi /data yang terpasang.',
        uninstall: 'Copot',
        uninstallConfirm: 'Hapus berkas Ventoy?',
        noImages: 'Tidak ada image untuk disk Ventoy.',
        onDisk: 'Di disk Ventoy',
        missing: 'Hilang: {{file}}',
        remove: 'Keluarkan dari disk Ventoy',
        setHint:
          'Kumpulan image hanya dapat diubah saat disk Ventoy tidak berada di drive mana pun.',
        useAsDisk: 'Gunakan sebagai disk virtual',
        failed: 'Permintaan Ventoy gagal',
        secureBoot:
          'Jika Secure Boot aktif, host harus mendaftarkan kunci Ventoy sekali di MokManager. Berkas kunci ENROLL_THIS_KEY_IN_MOKMANAGER.cer ada di partisi VTOYEFI.',
        readOnly:
          'Host melihat disk sebagai hanya-baca, jadi persistensi Ventoy dan ventoy.json di drive tidak berfungsi.'
      },
      tips: {
        title: 'Cara mengunggah',
        usb1: 'Hubungkan IronKVM ke komputer Anda melalui USB.',
        usb2: 'Pastikan disk virtual telah terpasang (Pengaturan - Disk Virtual).',
        usb3: 'Buka disk virtual di komputer Anda dan salin file gambar ke direktori root disk virtual.',
        scp1: 'Pastikan IronKVM dan komputer Anda berada di jaringan lokal yang sama.',
        scp2: 'Buka terminal di komputer Anda dan gunakan perintah SCP untuk mengunggah file gambar ke direktori /data di IronKVM.',
        scp3: 'Contoh: scp jalur-gambar-anda root@ip-nanokvm-anda:/data',
        tfCard: 'Kartu TF',
        tf1: 'Metode ini didukung di sistem linux',
        tf2: 'Dapatkan Kartu TF dari IronKVM (untuk versi LENGKAP, bongkar casingnya terlebih dahulu).',
        tf3: 'Masukkan Kartu TF ke pembaca kartu dan hubungkan ke komputer Anda.',
        tf4: 'Salin berkas gambar ke direktori /data pada Kartu TF.',
        tf5: 'Masukkan Kartu TF ke dalam IronKVM.'
      }
    },
    script: {
      title: 'Script',
      upload: 'Mengunggah',
      run: 'Jalankan',
      runBackground: 'Jalankan di belakang',
      runFailed: 'Gagal menjalankan',
      attention: 'Perhatian',
      delDesc: 'Apa kamu yakin menghapus data ini?',
      confirm: 'Ya',
      cancel: 'Tidak',
      delete: 'Hapus',
      close: 'Tutup',
      empty: 'Belum ada skrip. Unggah berkas .sh atau .py untuk menjalankannya di papan.',
      loadFailed: 'Gagal memuat skrip',
      uploaded: 'Skrip diunggah',
      uploadFailed: 'Gagal mengunggah skrip',
      started: 'Skrip dijalankan di latar belakang',
      deleteFailed: 'Gagal menghapus skrip',
      waitLimit: 'Menunggu skrip selesai, hingga {{minutes}} menit.',
      timedOut:
        'Skrip berjalan lebih dari {{minutes}} menit dan halaman ini berhenti menunggu. Skrip mungkin masih berjalan di papan.'
    },
    terminal: {
      invalidBaud: 'Baud rate ini tidak didukung.',
      invalidPort: 'Masukkan path perangkat di bawah /dev, misalnya /dev/ttyS1.',
      invalidSettings: 'Pengaturan port serial tidak valid. Ini adalah shell papan sendiri.',
      disconnected: 'Terputus. Tekan Enter untuk menyambung ulang.',
      title: 'Terminal',
      nanokvm: 'Terminal IronKVM',
      serial: 'Terminal Port Serial',
      serialPort: 'Port Serial',
      serialPortPlaceholder: 'Silahkan masukkan port serial',
      baudrate: 'Baud rate',
      parity: 'Paritas',
      parityNone: 'Tidak ada',
      parityEven: 'Genap',
      parityOdd: 'Ganjil',
      flowControl: 'Kontrol aliran',
      flowControlNone: 'Tidak ada',
      flowControlSoft: 'Perangkat lunak',
      flowControlHard: 'Perangkat keras',
      dataBits: 'Bit data',
      stopBits: 'Hentikan bit',
      confirm: 'Ok'
    },
    wol: {
      no: 'Tidak',
      yes: 'Ya',
      deleteConfirm: 'Hapus alamat tersimpan ini?',
      delete: 'Hapus',
      wake: 'Bangunkan',
      rename: 'Ganti nama',
      showMac: 'Tampilkan alamat MAC',
      showName: 'Tampilkan nama',
      requestFailed: 'Tidak dapat menjangkau perangkat untuk mengirim perintah',
      deleteFailed: 'Gagal menghapus',
      renameFailed: 'Gagal mengganti nama',
      title: 'Wake-on-LAN',
      sending: 'Kirim perintah...',
      sent: 'Perintah terkirim',
      input: 'Silahkan masukkan MAC',
      ok: 'Ok'
    },
    download: {
      uploadFailed: 'Unggahan gagal',
      uploadSuccess: 'Unggahan selesai',
      uploading: 'Mengunggah: {{file}}',
      downloadingPercent: 'Mengunduh ({{percent}}): {{file}}',
      downloading: 'Mengunduh: {{file}}',
      title: 'Pengunduh Gambar',
      input: 'Silakan masukkan gambar jarak jauh URL',
      ok: 'Ok',
      disabled: 'Partisi /data adalah RO, jadi kami tidak dapat mengunduh gambarnya',
      uploadbox: 'Letakkan file di sini atau klik untuk memilih',
      inputfile: 'Silakan masukkan File gambar',
      NoISO: 'Tidak ada ISO',
      sha256: 'SHA-256 (opsional)',
      sha256Placeholder: 'Masukkan checksum SHA-256 64 karakter',
      invalidSHA256: 'SHA-256 harus berupa string heksadesimal 64 karakter',
      failed: 'Unduhan gagal',
      success: 'Unduhan berhasil',
      checksumFailed: 'Unduhan gagal: verifikasi SHA-256 gagal',
      cancel: 'Batal',
      cancelFailed: 'Gagal membatalkan unduhan',
      bootMenu: 'Menu boot (netboot.xyz)',
      bootMenuPresent: '{{file}} sudah ada di perangkat, dengan checksum yang benar',
      bootMenuDesc: 'Unduh ISO netboot.xyz, checksum diperiksa, untuk CD virtual'
    },
    alerts: {
      title: 'Perlu perhatian',
      temperature: {
        warning: 'Papan bersuhu {{celsius}} °C. Pastikan udara dapat mencapainya.',
        critical: 'Papan bersuhu {{celsius}} °C, terlalu panas. Beri aliran udara atau matikan.'
      },
      storage: {
        warning:
          'Hanya {{available}} dari {{total}} yang kosong di {{path}}. Image besar mungkin tidak muat.',
        critical:
          'Hanya {{available}} yang kosong di {{path}}. Unggahan, unduhan, dan pemasangan add-on akan gagal. Hapus image yang tidak diperlukan.'
      },
      vpn: '{{name}} diatur untuk mulai saat boot tetapi tidak berjalan, jadi akses jarak jauh melaluinya terputus.',
      openVpn: 'Buka pengaturan VPN',
      stream: 'Stream video gagal. Coba mode video lain di menu Layar, atau muat ulang halaman.'
    },
    power: {
      resetDesc: 'Memulai ulang host seketika. Pekerjaan yang belum disimpan hilang.',
      powerShortDesc: 'Menyalakan host, atau meminta OS-nya untuk mati (ACPI).',
      powerLongDesc: 'Memaksa host mati tanpa shutdown.',
      hddLed: 'LED disk',
      hddActive: 'Aktif',
      hddIdle: 'Diam',
      title: 'Daya',
      showConfirm: 'Konfirmasi',
      showConfirmTip:
        'Tanyakan sebelum tekan singkat tombol daya. Reset dan tekan lama selalu bertanya.',
      reset: 'Mulai Ulang',
      power: 'Daya',
      powerShort: 'Data (tekan sebentar)',
      powerLong: 'Power (tekan lama)',
      resetConfirm: 'Lanjutkan operasi penyetelan ulang?',
      powerConfirm: 'Lanjutkan pengoperasian listrik?',
      okBtn: 'Ya',
      cancelBtn: 'Tidak',
      hostOs: 'OS host',
      hostOsTip: 'Dikirim sebagai tombol USB. Host yang menentukan fungsinya.',
      sleep: 'Tidur',
      wake: 'Bangunkan',
      wakeKey: 'Bangunkan dengan Shift',
      powerDown: 'Matikan',
      sleepConfirm: 'Tidurkan host?',
      powerDownConfirm: 'Kirim tombol matikan ke host?',
      wakeTip:
        'Host yang sedang tidur sering mengabaikan Bangunkan dari perangkat yang menidurkannya. Bangunkan dengan Shift menekan tombol keyboard, yang diterima lebih banyak host.',
      led: 'LED daya',
      ledOn: 'Menyala',
      ledOff: 'Mati',
      ledUnknown: 'Tidak diketahui',
      ledConnected: 'LED daya terhubung',
      ledConnectedTip:
        'Aktifkan hanya jika header LED daya host tersambung ke papan. Tanpa itu, status daya tidak diketahui.',
      ledConnectedFailed: 'Gagal menyimpan pengaturan LED daya',
      powerLongConfirm:
        'Tahan tombol daya selama {{seconds}} dtk? Ini memutus daya tanpa mematikan sistem.',
      done: 'Tombol ditekan',
      failed: 'Gagal menekan tombol'
    },
    settings: {
      title: 'Pengaturan',
      nav: {
        system: 'Sistem',
        network: 'Jaringan',
        access: 'Akses',
        integrations: 'Integrasi',
        boot: 'Boot dan media',
        browser: 'Browser ini',
        search: 'Cari pengaturan',
        noMatch: 'Tidak ada pengaturan yang cocok',
        locked:
          'Sebuah operasi sedang berjalan. Halaman lain dan tombol tutup tidak tersedia sampai selesai.',
        vpnProvider: 'Penyedia VPN'
      },
      mcp: {
        keyNote:
          'MCP memakai kunci API sendiri, ditampilkan di bawah. Kunci dari halaman Kunci API tidak berlaku di sini.',
        title: 'Layanan MCP',
        service: 'Kontrol jarak jauh MCP',
        serviceDesc:
          'Izinkan klien MCP tepercaya mengontrol keyboard dan mouse serta mengambil tangkapan layar',
        securityWarning:
          'Siapa pun yang memiliki kunci API ini dapat mengontrol host jarak jauh dan melihat layarnya. Gunakan HTTPS dan aktifkan hanya pada jaringan tepercaya.',
        endpoint: 'Endpoint',
        apiKey: 'Kunci API',
        regenerateConfirmTitle: 'Buat ulang kunci API MCP?',
        regenerateConfirmDesc: 'Kunci saat ini akan langsung berhenti berfungsi.',
        enableConfirmTitle: 'Aktifkan kontrol MCP eksternal?',
        enableConfirmDesc:
          'Mengaktifkan MCP akan menghentikan PicoClaw dan menutup semua sesi PicoClaw yang aktif.',
        failed: 'Operasi MCP gagal',
        copyFailed: 'Gagal menyalin. Salin secara manual.',
        okBtn: 'Konfirmasi',
        cancelBtn: 'Batal',
        showKey: 'Tampilkan kunci',
        hideKey: 'Sembunyikan kunci',
        regenerateKey: 'Buat ulang kunci'
      },
      redfish: {
        example: 'Contoh',
        title: 'Redfish',
        service: 'Layanan Redfish',
        serviceDesc:
          'API Redfish dari DMTF, untuk kontrol daya, media virtual, dan status dari alat seperti redfishtool dan Ansible. Menonaktifkannya akan mengakhiri semua sesi Redfish.',
        endpoint: 'Root layanan',
        httpsOn: 'Papan melayani HTTPS, yang dibutuhkan sebagian besar alat Redfish.',
        httpsOff:
          'Papan melayani HTTP biasa. Sebagian besar alat Redfish membutuhkan HTTPS: aktifkan di "Pengaturan > Jaringan".',
        credentials:
          'Redfish menerima akun KVM, dengan autentikasi Basic atau sesi Redfish, serta kunci API yang dikirim sebagai X-Auth-Token. Kunci API dikelola di halaman Kunci API.',
        powerActions: 'Tindakan daya',
        powerActionsDesc:
          'Jenis reset yang tersedia saat ini. On, ForceOff, dan GracefulShutdown memerlukan status daya, sehingga hanya tersedia jika "LED daya terhubung" diaktifkan di menu daya.',
        sessions: 'Sesi',
        noSessions: 'Tidak ada sesi Redfish yang terbuka',
        created: 'Dibuat',
        lastUsed: 'Terakhir digunakan',
        refresh: 'Segarkan',
        end: 'Akhiri',
        endConfirmTitle: 'Akhiri sesi Redfish ini?',
        endConfirmDesc: 'Tokennya langsung berhenti berfungsi. Klien harus masuk kembali.',
        failed: 'Operasi Redfish gagal',
        copyFailed: 'Gagal menyalin. Salin secara manual.',
        okBtn: 'Konfirmasi',
        cancelBtn: 'Batal'
      },
      ipmi: {
        copyBeforeSave:
          'Salin kata sandi sekarang. Setelah disimpan, kata sandi tidak dapat ditampilkan lagi.',
        noLogin:
          'IPMI aktif, tetapi tidak ada akun aktif yang memiliki kata sandi IPMI, sehingga tidak ada yang bisa masuk. Atur satu di bawah.',
        title: 'IPMI',
        warning:
          'Autentikasi IPMI memang lemah secara desain. Siapa pun yang dapat menjangkau board dan mengetahui nama pengguna dapat memperoleh hash kata sandi IPMI pengguna tersebut dan mencoba memecahkannya secara offline. Gunakan kata sandi yang dibuat otomatis, aktifkan IPMI hanya di jaringan tepercaya, dan utamakan Redfish lewat HTTPS jika alat mendukungnya.',
        service: 'IPMI melalui LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) pada port UDP 623, untuk daya dan status host. IPMI 1.5 dan cipher suite 0 ditolak. Menonaktifkannya mengakhiri semua sesi IPMI.',
        example: 'Contoh',
        copyFailed: 'Gagal menyalin. Salin secara manual.',
        ledOn: 'Status daya, on, off, soft, cycle, dan reset tersedia.',
        ledOff:
          '"LED daya terhubung" nonaktif di menu daya, sehingga status daya tidak diketahui. Hanya "power reset" yang berfungsi: status, on, off, soft, dan cycle ditolak.',
        accounts: 'Akun',
        accountsDesc:
          'IPMI masuk dengan akun KVM, masing-masing dengan kata sandi IPMI sendiri yang terpisah dari kata sandi web. Administrator mendapat ADMINISTRATOR. Pengguna mendapat USER: mereka dapat membaca status daya dengan "-L USER" tetapi tidak dapat mengubahnya.',
        passwordSet: 'Kata sandi IPMI sudah diatur',
        passwordNotSet: 'Tanpa kata sandi IPMI: tidak dapat masuk lewat IPMI',
        nameTooLong: 'Nama lebih dari 16 karakter, yang tidak diizinkan IPMI',
        accountDisabled: 'Akun dinonaktifkan',
        setPassword: 'Atur kata sandi',
        changePassword: 'Ubah kata sandi',
        remove: 'Hapus',
        removeConfirmTitle: 'Hapus kata sandi IPMI milik {{user}}?',
        removeConfirmDesc: 'Akun tidak dapat lagi masuk lewat IPMI, dan sesi IPMI-nya berakhir.',
        passwordTitle: 'Kata sandi IPMI untuk {{user}}',
        passwordDesc:
          '12 hingga 20 karakter ASCII yang dapat dicetak, berbeda dari kata sandi web. IPMI mengharuskan board menyimpan kata sandi dalam bentuk yang dapat dibaca kembali, jadi gunakan kata sandi yang tidak dipakai di tempat lain. Salin sebelum menyimpan: kata sandi tidak ditampilkan lagi.',
        passwordPlaceholder: 'Kata sandi IPMI',
        generate: 'Buat',
        copy: 'Salin',
        save: 'Simpan',
        passwordLength: 'Gunakan 12 hingga 20 karakter.',
        passwordChars: 'Gunakan hanya karakter ASCII yang dapat dicetak.',
        saved: 'Kata sandi IPMI disimpan',
        failed: 'Operasi IPMI gagal',
        okBtn: 'Konfirmasi',
        cancelBtn: 'Batal'
      },
      ssh: {
        service: 'Server SSH',
        serviceDesc: 'Jalankan sshd sekarang dan setiap kali boot',
        failed: 'Tidak dapat memuat pengaturan SSH',
        rootDefault: 'root masih memakai kata sandi pabrik',
        rootEmpty: 'root tidak punya kata sandi',
        rootWarning:
          'Siapa pun yang mencapai konsol atau SSH dapat masuk sebagai root. Atur kata sandi di {{account}} > {{password}}: untuk pemilik perangkat, ini juga mengatur kata sandi root.',
        connection: 'Koneksi',
        command: 'Masuk sebagai root',
        port: 'Port',
        viaVpn: 'Melalui {{name}}',
        notRunning: 'sshd tidak berjalan. Nyalakan server SSH untuk terhubung.',
        hostKeys: 'Sidik jari kunci host',
        hostKeysDesc: 'Cocokkan dengan yang ditampilkan ssh saat pertama kali terhubung.',
        noHostKeys: 'Belum ada kunci host. sshd membuatnya saat pertama kali berjalan.',
        keys: 'Kunci resmi',
        keysDesc:
          'Kunci publik yang dapat masuk sebagai root. Disimpan di partisi data, jadi pembaruan tetap menyimpannya.',
        noKeys: 'Belum ada kunci resmi.',
        noComment: 'tanpa komentar',
        addPlaceholder: 'Tempel satu kunci publik, misalnya isi ~/.ssh/id_ed25519.pub',
        add: 'Tambah kunci',
        added: 'Kunci ditambahkan',
        removed: 'Kunci dihapus',
        deleteConfirm: 'Hapus kunci ini?',
        deleteConfirmDesc: 'Kunci ini tidak bisa lagi masuk. Sesi yang terbuka tetap terbuka.',
        invalidKey: 'Ini bukan kunci publik. Tempel satu baris dari berkas .pub.',
        keyOptions: 'Kunci dengan opsi seperti command= atau from= tidak diterima di sini.',
        duplicateKey: 'Kunci ini sudah diizinkan.',
        lastKey: 'Kunci terakhir tidak dapat dihapus selama login hanya dengan kunci aktif.',
        keysOnly: 'Hanya kunci',
        keysOnlyDesc:
          'Matikan login dengan kata sandi dan keyboard-interactive. Sesi yang terbuka tetap terbuka.',
        keysOnlyNeedsKey: 'Tambahkan kunci resmi dulu, atau tidak ada yang bisa masuk.',
        keysOnlyOn: 'Login dengan kata sandi dimatikan',
        keysOnlyOff: 'Login dengan kata sandi dinyalakan',
        notHonoured:
          'sshd pada image ini tidak membaca pengaturan ini, jadi login dengan kata sandi tetap aktif.',
        reloadFailed:
          'Tersimpan, tetapi sshd tidak dapat dimuat ulang. Berlaku saat sshd berikutnya dijalankan.',
        notApplied:
          'sshd masih menerima kata sandi. Matikan lalu nyalakan server SSH untuk menerapkan pengaturan.',
        changePort: 'Ubah',
        portConfirm: 'Ubah port SSH menjadi {{port}}?',
        portConfirmDesc:
          'Sesi SSH Anda yang sedang berjalan tetap terbuka. Koneksi baru harus memakai port {{port}}. Pastikan firewall Anda mengizinkannya.',
        portChanged: 'Port SSH diubah menjadi {{port}}',
        portInvalid: 'Masukkan port dari 1 sampai 65535.',
        portReserved: 'IronKVM sendiri memakai port ini. Pilih port lain.',
        portInUse: 'Program lain di IronKVM sudah mendengarkan di port ini.',
        portNotHonoured:
          'sshd pada image ini tidak membaca pengaturan ini, jadi port tetap seperti semula.'
      },
      vnc: {
        address: 'Alamat',
        certHint:
          'VeNCrypt X509Plain memakai sertifikat swatanda perangkat, jadi klien memberi peringatan saat koneksi pertama. Terima saja, atau simpan sertifikat dari alamat HTTPS halaman ini dan berikan ke TigerVNC dengan -X509CA=<file>.',
        title: 'VNC',
        service: 'Server VNC',
        serviceDesc:
          'Memungkinkan klien VNC, seperti TigerVNC atau Remmina, melihat dan mengendalikan host. Klien harus mendukung encoding Tight. Satu sesi pada satu waktu.',
        credentials:
          'Masuk dengan akun KVM. Koneksi dienkripsi dengan sertifikat TLS papan (VeNCrypt X509Plain).',
        port: 'Port',
        portDesc: 'Port TCP tempat server mendengarkan.',
        maxFps: 'Batas frame rate',
        maxFpsDesc: 'Jumlah frame per detik terbanyak yang dikirim ke klien.',
        vncAuth: 'Autentikasi VNC biasa',
        vncAuthDesc:
          'Untuk klien tanpa VeNCrypt. Autentikasi ini memeriksa kata sandi VNC tersendiri, bukan akun.',
        vncAuthWarning:
          'Autentikasi VNC biasa tidak mengenkripsi koneksi. Siapa pun di jalur jaringan dapat melihat layar dan ketikan. Gunakan hanya di jaringan tepercaya.',
        password: 'Kata sandi VNC',
        passwordSet: 'Kata sandi sudah diatur. Ketik yang baru untuk mengubahnya.',
        passwordInvalid: 'Kata sandi VNC harus 6 sampai 8 karakter.',
        save: 'Simpan',
        saved: 'Pengaturan disimpan',
        state: 'Status',
        listening: 'Mendengarkan di port {{port}}',
        notListening: 'Tidak mendengarkan',
        noSession: 'Tidak ada sesi terbuka',
        client: 'Klien',
        user: 'Pengguna',
        method: 'Autentikasi',
        methodVencrypt: 'Akun melalui TLS',
        methodVnc: 'Kata sandi VNC',
        since: 'Terhubung sejak',
        resolution: 'Resolusi',
        framesSent: 'Frame terkirim',
        lastError: 'Sesi terakhir berakhir: {{error}}',
        refresh: 'Segarkan',
        disconnect: 'Putuskan',
        disconnectConfirmTitle: 'Akhiri sesi VNC?',
        disconnectConfirmDesc:
          'Klien langsung diputus, dan setiap tombol serta tombol mouse yang ditahannya dilepas.',
        failed: 'Operasi VNC gagal',
        okBtn: 'Konfirmasi',
        cancelBtn: 'Batal'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Watchdog host',
        serviceDesc:
          'Jika host seharusnya menyala dan gambarnya tidak berubah, atau tidak ada sinyal HDMI, selama batas waktu, papan menekan reset atau mematikan lalu menyalakan host.',
        stillWarning:
          'Host yang layarnya masuk mode tidur, atau yang gambarnya diam saat bekerja, tampak macet. Matikan mode tidur layar di host, atau atur alamat ping.',
        ledHint:
          '"LED daya terhubung" dimatikan di menu daya. Watchdog tidak bisa melihat kapan host mati, jadi menganggap host selalu menyala.',
        timeout: 'Batas waktu',
        timeoutDesc:
          'Berapa lama host boleh tidak menunjukkan tanda kehidupan sebelum watchdog bertindak.',
        action: 'Tindakan',
        actionDesc:
          'Matikan lalu nyalakan menahan tombol daya selama 5 detik, lalu menekannya lagi.',
        actionReset: 'Mulai Ulang',
        actionPower: 'Matikan lalu nyalakan',
        cooldown: 'Jeda',
        cooldownDesc: 'Waktu terpendek di antara dua tindakan.',
        maxPerHour: 'Tindakan per jam',
        maxPerHourDesc: 'Jumlah tindakan terbanyak dalam satu jam.',
        pingHost: 'Alamat ping',
        pingHostDesc:
          'Alamat IP host. Balasan dihitung sebagai tanda kehidupan. Kosongkan agar tidak melakukan ping.',
        pingHostInvalid: 'Masukkan alamat IPv4 atau IPv6.',
        minutes: 'menit',
        save: 'Simpan',
        saved: 'Tersimpan',
        state: 'Detektor',
        status: {
          off: 'Mati',
          watching: 'Mengawasi',
          hostOff: 'Host mati',
          captureOff: 'Tangkapan HDMI mati',
          cooldown: 'Jeda',
          capped: 'Batas per jam tercapai',
          acting: 'Bertindak'
        },
        signal: 'Sinyal HDMI',
        yes: 'Ya',
        no: 'Tidak',
        led: 'LED daya',
        on: 'Menyala',
        off: 'Mati',
        ledNotConnected: 'Tidak terhubung',
        ping: 'Ping',
        pingNotSet: 'Belum diatur',
        pingReply: 'Membalas',
        pingNoReply: 'Tidak ada balasan',
        lastChange: 'Perubahan gambar terakhir',
        never: 'Belum pernah',
        actsIn: 'Bertindak dalam',
        actionsLastHour: 'Tindakan dalam satu jam terakhir',
        duration: '{{minutes}} menit {{seconds}} detik',
        log: 'Log',
        noLog: 'Watchdog belum pernah bertindak.',
        refresh: 'Muat ulang',
        reasonFrozen: 'Gambar tidak berubah',
        reasonNoSignal: 'Tidak ada sinyal HDMI',
        stuckFor: 'tidak ada tanda kehidupan selama {{duration}}',
        pressFailed: 'Penekanan gagal: {{error}}',
        noScreenshot: 'Tidak ada tangkapan layar',
        failed: 'Operasi watchdog gagal',
        powerNeedsLed: 'Siklus daya memerlukan "LED daya terhubung" di menu daya.',
        noLedConfirmTitle: 'Aktifkan watchdog tanpa LED daya?',
        noLedConfirmDesc:
          'Papan tidak dapat melihat kapan host mati, sehingga host dianggap selalu menyala. Jika Anda mematikan host, watchdog menekan reset setelah batas waktu lewat. Hubungkan LED daya untuk menghindarinya.',
        noLedConfirmOk: 'Aktifkan',
        cancel: 'Batal'
      },
      media: {
        title: 'Media virtual',
        description:
          'Penyiapan untuk dialog Media di bilah alat. Memasang image, menambahkannya, dan memilih set Ventoy dilakukan di dialog.',
        ejectFirst:
          'Disk Ventoy ada di dalam drive. Keluarkan di dialog Media untuk mencopot pemasangan.'
      },
      netboot: {
        title: 'Boot jaringan',
        isoDownload: 'Unduh',
        description:
          'Boot host dari jaringan: iPXE dan menu image di KVM lewat tautan jaringan USB, atau netboot.xyz lewat proxy DHCP di LAN.',
        addon: 'dnsmasq dan file boot',
        addonDesc:
          'Dipasang di /data: dnsmasq dari Alpine, iPXE dan netboot.xyz dari rilisnya, masing-masing diperiksa dengan checksum-nya.',
        install: 'Pasang',
        installing: 'Sedang memasang. Ini bisa memakan waktu beberapa menit.',
        uninstall: 'Copot',
        uninstallConfirm: 'Matikan boot jaringan dan hapus dnsmasq serta file boot?',
        needsData: 'Boot jaringan memerlukan image IronKVM dengan partisi /data terpasang.',
        usb: 'Di tautan jaringan USB',
        usbDesc:
          'Selama tautan jaringan USB aktif, dnsmasq melayaninya menggantikan udhcpd. Host mendapat satu alamatnya tanpa router dan tanpa server DNS, iPXE untuk arsitekturnya, dan menu image ISO di KVM.',
        linkOff: 'Tautan jaringan USB mati. Nyalakan di Perangkat, Jaringan USB.',
        menuUrl: 'Menu',
        leases: 'Lease host',
        noLeases: 'Belum ada',
        netbootxyzNote:
          'netboot.xyz di menu dimuat dari internet, yang tidak dijangkau tautan USB. Host memerlukan internet di port jaringan lain.',
        lan: 'Proxy DHCP di LAN',
        lanDesc:
          'Menjawab klien PXE di LAN dengan netboot.xyz, yang lalu memuat menunya dari internet. Tidak pernah membagikan alamat dan tidak menyajikan image di KVM.',
        lanWarning:
          'Setiap klien PXE di LAN ini ditawari netboot.xyz, bukan hanya host. Nyalakan ini hanya di jaringan yang Anda kendalikan.',
        lanConfirm: 'Nyalakan proxy DHCP di LAN?',
        lanInterface: 'LAN',
        running: 'Berjalan',
        stopped: 'Tidak berjalan',
        images: 'Image di menu',
        noImages: 'Tidak ada image ISO di direktori image.',
        boots: 'Boot terbaru',
        noBoots: 'Host belum mengambil apa pun.',
        log: 'Log dnsmasq',
        refresh: 'Muat ulang',
        okBtn: 'Konfirmasi',
        cancelBtn: 'Batal',
        failed: 'Operasi boot jaringan gagal'
      },
      about: {
        title: 'Tentang IronKVM',
        information: 'Informasi',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versi Aplikasi',
        applicationTip: 'Versi aplikasi web IronKVM',
        image: 'Version Gambar',
        imageTip: 'Image kartu IronKVM, dan image sistem NanoKVM yang menjadi dasarnya',
        kernel: 'Versi Kernel',
        kernelTip: 'Rilis kernel Linux yang sedang berjalan',
        deviceKey: 'Kunci Perangkat',
        videoMemory: 'Memori Video',
        videoMemoryTip:
          'Memori yang dicadangkan untuk penangkapan video. Memori ini tidak dibagi dengan bagian sistem lainnya.',
        videoMemoryGenerations_other:
          '{{count}} sesi IronKVM sebelumnya masih menahan memori video',
        videoMemoryReboot: 'Mulai ulang untuk mengambilnya kembali.',
        community: 'Komunitas',
        hostname: 'Nama Host',
        hostnameUpdated: 'Nama host diperbarui. Nyalakan ulang untuk menerapkan.',
        ipType: {
          Wired: 'Berkabel',
          Wireless: 'Nirkabel',
          Other: 'Lainnya'
        },
        hostnameInvalid:
          'Gunakan huruf, angka, dan tanda hubung, maksimal 63 per bagian yang dipisah titik. Tanpa tanda hubung di awal atau akhir bagian.',
        hostnameFailed: 'Gagal mengubah nama host',
        editHostname: 'Ubah nama host',
        docs: 'Dokumentasi',
        hardware: 'Perangkat keras',
        hardwareFaq: 'FAQ perangkat keras',
        disclaimer:
          'IronKVM: firmware komunitas yang diperkuat untuk Sipeed NanoKVM. Tidak berafiliasi dengan Sipeed.',
        basedOn: 'berbasis NanoKVM {{version}}'
      },
      preferences: {
        title: 'Preferensi'
      },
      performance: {
        title: 'Performa'
      },
      appearance: {
        thisBrowser: 'Browser ini',
        thisBrowserDesc:
          'Hanya disimpan di browser ini. Browser lain menyimpan pengaturannya sendiri.',
        deviceWide: 'Perangkat',
        deviceWideDesc: 'Disimpan di perangkat. Berlaku untuk semua orang yang membukanya.',
        language: 'Bahasa',
        languageDesc: 'Pilih bahasa untuk antarmuka',
        webTitle: 'Judul Web',
        webTitleDesc: 'Menyesuaikan judul halaman web',
        menuBar: {
          title: 'Bilah Menu',
          mode: 'Mode Tampilan',
          modeDesc: 'Menampilkan bilah menu di layar',
          modeOff: 'Mati',
          modeAuto: 'Sembunyikan otomatis',
          modeAlways: 'Selalu terlihat',
          keyboardLedStatus: 'Indikator kunci keyboard',
          keyboardLedStatusDesc:
            'Tampilkan status Num Lock, Caps Lock, dan Scroll Lock komputer jarak jauh',
          icons: 'Ikon Submenu',
          iconsDesc: 'Menampilkan ikon submenu di bilah menu'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Status kunci keyboard jarak jauh',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Aktif',
        off: 'Nonaktif',
        unknown: 'Tidak diketahui'
      },
      device: {
        title: 'Perangkat',
        oled: {
          title: 'OLED',
          description: 'OLED screen automatically sleep',
          brightness: 'Kecerahan OLED',
          brightnessDescription: 'Tingkat yang lebih rendah membuat layar lebih awet',
          brightnessLevels: {
            '64': 'Terendah',
            '96': 'Rendah',
            '128': 'Sedang',
            '160': 'Tinggi',
            '207': 'Default',
            '255': 'Maksimum'
          },
          0: 'Tidak pernah',
          15: '15 detik',
          30: '30 detik',
          60: '1 menit',
          180: '3 menit',
          300: '5 menit',
          600: '10 menit',
          1800: '30 menit',
          3600: '1 jam'
        },
        sections: {
          video: 'Video',
          usb: 'USB',
          frontPanel: 'Panel depan'
        },
        cpuFreq: {
          title: 'Frekuensi CPU',
          description: 'Atur clock CPU yang diterapkan pada boot berikutnya',
          tip: 'CPU melakukan boot pada 850 MHz dan dirancang untuk 1000 MHz. Nilai baru diterapkan pada boot berikutnya, bukan saat sistem berjalan. 1000 MHz masih sesuai spesifikasi; suhu tetap jauh di bawah batas pada kedua pengaturan.',
          running: 'Berjalan: {{mhz}} MHz',
          rebootToApply: 'mulai ulang untuk menerapkan',
          rebootConfirm: 'Mulai ulang sekarang untuk menerapkan {{mhz}} MHz?'
        },
        swap: {
          title: 'Tukar',
          disable: 'Nonaktifkan',
          description: 'Atur ukuran file swap',
          tip: 'Mengaktifkan fitur ini dapat mempersingkat masa pakai kartu SD Anda!',
          active: 'Aktif - {{used}} dari {{total}}',
          inactive: 'Diatur, tetapi tidak dipakai'
        },
        zram: {
          title: 'Swap terkompresi (zram)',
          description: 'Swap di RAM terkompresi, bukan di kartu SD',
          tip: 'zram menjauhkan swap dari kartu SD, sehingga tidak menyebabkan keausan. Tidak ada swap disk di belakangnya: jika zram penuh, kernel menghentikan sebuah proses alih-alih melakukan paging secara lambat. Batas memori membatasi seberapa banyak RAM yang dapat dipakai zram.',
          unavailable: 'Modul kernel tidak terpasang di perangkat ini',
          inactive: 'Diaktifkan, tetapi perangkat tidak berjalan',
          active: 'Aktif - {{used}} dari {{total}}, {{ratio}}x',
          off: 'Mati',
          detail: {
            algorithm: 'Algoritma: {{algorithm}}',
            memory: 'Memori terpakai: {{used}} dari {{limit}}',
            memoryNoLimit: 'Memori terpakai: {{used}}, tanpa batas',
            counters: 'Halaman swap masuk {{in}}, keluar {{out}} (semua perangkat swap, sejak boot)'
          }
        },
        mouseJiggler: {
          title: 'Tikus Jiggler',
          description: 'Mencegah host jarak jauh tertidur',
          disable: 'Nonaktifkan',
          absolute: 'Mode Absolut',
          relative: 'Mode Relatif'
        },
        mdns: {
          description: 'Aktifkan layanan penemuan mDNS',
          tip: 'Mematikan jika tidak diperlukan'
        },
        hdmi: {
          description: 'Aktifkan keluaran HDMI/monitor',
          idleTimeoutTitle: 'Batas waktu tangkapan tidak aktif',
          idleTimeoutDescription: 'Hentikan tangkapan HDMI setelah tidak ada penonton aktif selama',
          minutes: 'mnt'
        },
        hidOnly: 'HID-Mode Hanya',
        hidOnlyDesc: 'Berhenti meniru perangkat virtual, hanya mempertahankan kontrol dasar HID',
        disk: 'Disk virtual',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Jaringan virtual',
        networkDesc: 'Pasang kartu jaringan virtual pada host jarak jauh',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Host:',
          description:
            'Tautan jaringan privat ke host jarak jauh melalui kabel USB. Host mendapat alamat tanpa gateway dan tanpa DNS, sehingga tidak dapat menjangkau LAN Anda melalui IronKVM.',
          mode: 'Protokol',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (untuk host tanpa NCM)',
          rndis: 'RNDIS (tidak lagi ditawarkan)',
          rndisNote: 'Tautan ini memakai RNDIS, yang tidak lagi ditawarkan. Pilih NCM atau ECM.',
          subnet: 'Subnet',
          subnetDesc:
            'Jaringan IPv4 privat, /24 hingga /30. IronKVM memakai alamat pertama, host memakai alamat kedua.',
          invalidSubnet: 'Masukkan subnet seperti 172.31.255.0/30.',
          apply: 'Terapkan',
          confirm: 'Sambungkan ulang perangkat USB?',
          reenumerate:
            'Menerapkan akan membangun ulang koneksi USB. Host kehilangan keyboard, mouse, dan disk virtual selama beberapa detik.'
        },
        audio: 'Speaker Virtual',
        audioDesc:
          'Menyediakan kartu suara USB untuk host jarak jauh, sehingga Anda dapat mendengarnya. Host harus memilihnya sebagai perangkat output. Mengubah ini akan membangun ulang koneksi USB.',
        audioNote: 'Audio tersedia di kedua mode H.264 (WebRTC dan Direct), tidak di MJPEG',
        console: 'Konsol Serial',
        consoleDesc:
          'Menyediakan port serial USB untuk host jarak jauh, untuk masuk ke IronKVM ini saat jaringan tidak dapat dijangkau',
        consoleTip:
          'Siapa pun yang mengendalikan host jarak jauh akan mendapatkan prompt login IronKVM ini. Tetapkan kata sandi yang kuat sebelum mengaktifkan (Akun - Ubah Kata Sandi).',
        usbApply: {
          changed: 'Diubah',
          discard: 'Buang',
          pending: 'Perubahan belum diterapkan.'
        },
        endpoints: {
          title: 'Slot USB',
          free: '{{free}} dari {{total}} kosong',
          slots: 'Slot: {{count}}',
          full: 'Slot USB kosong tidak cukup. Matikan hal lain terlebih dahulu.',
          inactive: 'Aktif, tetapi tidak berjalan: pengontrol USB kehabisan slot. Matikan perangkat lain dan perangkat ini langsung berjalan.',
          explain: 'Pengontrol USB memiliki jumlah slot (endpoint masuk) yang tetap, dan keyboard serta mouse selalu memakai sebagian. Jika perangkat yang aktif lebih banyak dari yang muat, keyboard dan mouse dipertahankan dan sisanya dimatikan.',
          error: 'Tidak dapat menjangkau perangkat. Coba lagi.',
          fitTogether: 'Yang muat bersamaan: {{sets}}'
        },
        reboot: 'Mulai ulang',
        rebootDesc: 'Apakah Anda yakin ingin me-reboot IronKVM?',
        okBtn: 'Ya',
        cancelBtn: 'Tidak',
        rebootFailed: 'Mulai ulang gagal'
      },
      network: {
        title: 'Jaringan',
        wifi: {
          disconnectBtn: 'Putuskan',
          disconnectWarning:
            'Jika Anda mengakses IronKVM melalui jaringan Wi-Fi ini, halaman ini akan kehilangan koneksi.',
          disconnected: 'Wi-Fi terputus',
          title: 'Wi-Fi',
          description: 'Konfigurasi Wi-Fi',
          apMode: 'Mode AP aktif, sambungkan ke Wi-Fi dengan memindai kode QR',
          connect: 'Hubungkan Wi-Fi',
          connectDesc1: 'Masukkan SSID jaringan dan kata sandi',
          connectDesc2: 'Masukkan kata sandi untuk bergabung ke jaringan ini',
          disconnect: 'Yakin ingin memutuskan jaringan?',
          failed: 'Koneksi gagal, coba lagi.',
          ssid: 'Nama',
          password: 'Kata sandi',
          joinBtn: 'Gabung',
          confirmBtn: 'OK',
          cancelBtn: 'Batal'
        },
        tls: {
          description: 'Aktifkan protokol HTTPS',
          tip: 'Perhatian: Menggunakan HTTPS dapat meningkatkan latensi, terutama pada mode video MJPEG.',
          restarting: 'Memulai ulang server perangkat, ini memakan waktu sekitar dua menit...',
          waiting: 'Menunggu perangkat merespons kembali...',
          waitingHttp:
            'Beralih kembali ke http. Muat ulang halaman ini jika tidak terbuka dengan sendirinya.',
          failed: 'Tidak dapat mengubah pengaturan HTTPS',
          enableConfirm: 'Aktifkan HTTPS?',
          disableConfirm: 'Nonaktifkan HTTPS?',
          confirmDesc:
            'Ini mengeluarkan Anda dan memulai ulang server perangkat, yang memakan waktu sekitar dua menit. Halaman lalu membuka {{url}}.',
          confirmOk: 'Lanjutkan',
          confirmCancel: 'Batal'
        },
        ethernet: {
          title: 'Alamat IP',
          description: 'Atur cara IronKVM memperoleh alamatnya di jaringan kabel',
          dhcp: 'DHCP',
          manual: 'Manual',
          networkDetails: 'Detail Jaringan',
          interface: 'Antarmuka',
          ipAddress: 'Alamat IP',
          subnetMask: 'Subnet Mask',
          router: 'Router',
          save: 'Terapkan',
          invalidAddress: 'Masukkan alamat IP yang valid',
          invalidMask: 'Masukkan subnet mask yang valid, misalnya 255.255.255.0 atau 24',
          invalidRouter: 'Masukkan alamat router yang valid',
          addressRequired: 'Alamat IP wajib diisi',
          maskRequired: 'Subnet mask wajib diisi',
          applyTitle: 'Ubah alamat IronKVM?',
          applyWarning:
            'Koneksi ke halaman ini akan terputus. IronKVM menerapkan alamat baru dan menunggu {{seconds}} detik sampai Anda menjangkaunya di sana. Menjangkaunya akan mempertahankan perubahan. Jika tidak ada yang menjangkaunya, IronKVM mengembalikan pengaturan sebelumnya.',
          applyConfirm: 'Terapkan',
          applyCancel: 'Batal',
          applyFailed: 'Gagal menerapkan alamat',
          trialTitle: 'Menunggu konfirmasi',
          trialDhcp: 'IronKVM sedang meminta alamat melalui DHCP.',
          trialStatic: 'IronKVM sekarang berada di {{address}}.',
          trialInstruction:
            'Buka IronKVM di alamat barunya dan masuk jika diminta. Menjangkaunya di sana akan mempertahankan perubahan. Jika tidak ada yang menjangkau IronKVM dalam {{seconds}} detik, IronKVM mengembalikan pengaturan sebelumnya.',
          trialOpen: 'Buka alamat baru',
          trialKeep: 'Pertahankan pengaturan ini',
          trialKept: 'Alamat baru tersimpan',
          trialKeepFailed: 'Gagal mempertahankan pengaturan',
          trialGone: 'Perubahan sudah dikembalikan. Coba lagi.',
          unsaved: 'Perubahan belum disimpan'
        },
        dns: {
          title: 'DNS',
          description: 'Konfigurasi server DNS untuk IronKVM',
          mode: 'Mode',
          dhcp: 'DHCP',
          manual: 'Manual',
          add: 'Tambah DNS',
          save: 'Simpan',
          invalid: 'Masukkan alamat IP yang valid',
          noDhcp: 'DNS DHCP saat ini tidak tersedia',
          saved: 'Pengaturan DNS disimpan',
          saveFailed: 'Gagal menyimpan pengaturan DNS',
          unsaved: 'Perubahan belum disimpan',
          maxServers: 'Maksimal {{count}} server DNS diizinkan',
          dnsServers: 'Server DNS',
          dhcpServersDescription: 'Server DNS diperoleh otomatis dari DHCP',
          manualServersDescription: 'Server DNS dapat diedit secara manual',
          networkDetails: 'Detail Jaringan',
          interface: 'Antarmuka',
          ipAddress: 'Alamat IP',
          subnetMask: 'Subnet mask',
          router: 'Router',
          none: 'Tidak ada'
        }
      },
      vpn: {
        connect: 'Hubungkan',
        connectDesc:
          'Bergabung ke jaringan {{name}}. Mati memutus koneksi tanpa menghentikan layanan.',
        kvmUrl: 'Alamat KVM',
        moreTip: 'Tindakan lain',
        restartTip: 'Mulai ulang',
        stopTip: 'Hentikan',
        updateTip: 'Perbarui ke {{version}}',
        loading: 'Memuat...',
        okBtn: 'Ya',
        cancelBtn: 'Tidak',
        restart: 'Mulai ulang {{name}}?',
        stop: 'Hentikan {{name}}?',
        stopDesc:
          'Daemon berhenti sekarang. Mulai saat boot adalah pengaturan terpisah dan tetap seperti semula.',
        update: 'Perbarui {{name}} ke {{version}}?',
        updateDesc: 'Daemon dimulai ulang jika sedang berjalan. Login tetap tersimpan.',
        notInstall: '{{name}} belum terinstal.',
        install: 'Instal',
        installing: 'Menginstal',
        installFailed: 'Instalasi gagal',
        retry: 'Coba lagi',
        notRunning: '{{name}} tidak berjalan. Jalankan untuk melanjutkan.',
        run: 'Mulai',
        boot: 'Mulai saat boot',
        bootDesc: 'Jalankan {{name}} saat KVM melakukan boot.',
        control: 'Server kontrol',
        connected: 'Terhubung',
        disconnected: 'Tidak terhubung',
        deviceName: 'Nama perangkat',
        deviceIP: 'IP perangkat',
        account: 'Akun',
        version: 'Versi',
        uptime: 'Waktu aktif',
        peers: 'Peer',
        noPeers: 'Belum ada peer.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Memori',
        daemonRss: 'Daemon',
        group: 'Grup add-on',
        high: 'diperlambat di atas {{size}}',
        max: 'dihentikan oleh kernel di atas {{size}}',
        noGroup: 'Tidak ada grup memori add-on di papan ini.',
        uninstall: 'Copot {{name}}',
        uninstallDesc: 'Apakah Anda yakin ingin mencopot {{name}}? Login tetap tersimpan di papan.',
        blocked:
          '{{other}} sedang berjalan atau dimulai saat boot. Hanya satu VPN yang dapat berjalan dalam satu waktu: hentikan {{other}} dan nonaktifkan mulai saat boot-nya terlebih dahulu.',
        swap: {
          title: 'Memori swap',
          tip: 'Jika daemon kekurangan memori, coba aktifkan swap. Swap diatur di "Pengaturan > Performa".'
        },
        copy: 'Salin',
        copied: 'Tautan disalin',
        copyFailed: 'Tidak dapat menyalin tautan. Pilih lalu salin secara manual.',
        open: 'Buka',
        checkAgain: 'Periksa lagi',
        notSignedIn: 'Belum masuk. Selesaikan masuk melalui tautan, lalu periksa lagi.',
        checkFailed: 'Tidak dapat memeriksa status masuk',
        loginWaiting: 'Halaman ini memeriksa setiap beberapa detik dan lanjut setelah Anda masuk.',
        uninstallFailed: 'Gagal menghapus instalasi',
        loginFailed: 'Gagal masuk'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Mengunduh',
        package: 'paket instalasi',
        unzip: 'dan unzip itu',
        notLogin:
          'Perangkat belum ditautkan. Silakan masuk dan tautkan perangkat ini ke akun Anda.',
        urlPeriod: 'Url ini berlaku selama 10 menit',
        login: 'Masuk',
        logout: 'Keluar',
        logoutDesc: 'Apakah Anda yakin ingin logout?',
        manualIntro: 'Atau instal secara manual lewat SSH:',
        copyBinaries: 'Salin tailscale dan tailscaled ke {{dir}} di IronKVM',
        linksFile: 'Di direktori yang sama, buat file bernama links berisi dua baris ini:',
        rebootRefresh: 'Mulai ulang IronKVM, lalu muat ulang halaman ini'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Perangkat ini belum bergabung ke jaringan NetBird. Bergabunglah dengan setup key, atau masuk dengan SSO.',
        setupKey: 'Kunci penyiapan',
        setupKeyPlaceholder: 'Tempel setup key dari dasbor NetBird',
        join: 'Gabung',
        or: 'atau',
        sso: 'Masuk dengan SSO',
        urlPeriod: 'Url ini berlaku selama 10 menit',
        logout: 'Batalkan pendaftaran',
        logoutDesc:
          'Membatalkan pendaftaran akan menghapus peer ini dari akun NetBird Anda dan menghapus konfigurasinya di sini. Untuk bergabung lagi diperlukan setup key atau login SSO, dan peer mungkin mendapat IP baru. Lanjutkan?',
        joinFailed: 'Tidak dapat bergabung ke jaringan'
      },
      update: {
        title: 'Periksa pembaruan',
        queryFailed: 'Gagal mendapatkan versi',
        updateFailed: 'Gagal memperbarui, tolong coba lagi.',
        isLatest: 'Kamu sudah menggunakan versi terbaru.',
        available: 'Ada pembaruan baru. apa kamu mau memperbarui?',
        updating: 'Pembaruan dimulai. Silahkan tunggu...',
        confirm: 'Konfirmasi',
        cancel: 'Batalkan',
        preview: 'Pratinjau Pembaruan',
        previewDesc: 'Dapatkan akses awal ke fitur dan peningkatan baru',
        previewTip:
          'Perlu diketahui bahwa rilis pratinjau mungkin mengandung bug atau fungsi yang tidak lengkap!',
        customServer: {
          title: 'Server Pembaruan Kustom',
          desc: 'Periksa dan unduh pembaruan daring dari server yang ditentukan',
          invalidUrl:
            'Masukkan direktori server HTTP atau HTTPS yang valid tanpa kueri, fragmen, atau latest.json.',
          loadFailed: 'Gagal memuat konfigurasi server pembaruan.',
          saveFailed: 'Gagal menyimpan konfigurasi server pembaruan.',
          saved: 'Konfigurasi server pembaruan telah disimpan.',
          save: 'Simpan',
          confirmTitle: 'Gunakan server pembaruan kustom?',
          confirmDesc:
            'SHA-512 hanya memeriksa bahwa paket cocok dengan manifes yang disediakan oleh server ini. Pemeriksaan ini tidak membuktikan bahwa paket tersebut merupakan rilis resmi IronKVM. Server yang bermasalah atau berbahaya dapat membuat perangkat tidak dapat digunakan, menyebabkan kehilangan data, atau membahayakan sistem.',
          confirm: 'Tetap Gunakan',
          useSipeed: 'Gunakan server resmi Sipeed',
          previewDisabled:
            'Pembaruan Pratinjau tidak tersedia saat server pembaruan kustom diaktifkan.'
        },
        offline: {
          chooseFile: 'Pilih file',
          installing: 'Unggahan selesai. Memasang...',
          noFile: 'Belum ada file dipilih',
          title: 'Pembaruan Offline',
          desc: 'Perbarui melalui paket instalasi lokal',
          upload: 'Mengunggah',
          checksumPlaceholder: 'Checksum SHA-256 (opsional)',
          invalidChecksum: 'Checksum SHA-256 harus berisi 64 karakter heksadesimal.',
          checksumMismatch: 'Verifikasi SHA-256 gagal. Paket mungkin rusak.',
          invalidName: 'Format nama file tidak valid. Silakan unduh dari rilis GitHub.',
          updateFailed: 'Gagal memperbarui, tolong coba lagi.'
        },
        updateTo: 'Perbarui ke {{version}}',
        updateConfirmDesc:
          'Perangkat memasang pembaruan dan memulai ulang servernya. Halaman ini dimuat ulang saat server kembali.',
        releaseNotes: 'Catatan rilis'
      },
      account: {
        title: 'Akun',
        webAccount: 'Nama akun web',
        role: 'Peran',
        roles: { admin: 'Administrator', user: 'Pengguna' },
        password: 'Kata sandi',
        updateBtn: 'Update',
        logoutBtn: 'Keluar',
        logoutDesc: 'Apakah Anda yakin ingin logout?',
        okBtn: 'Ya',
        cancelBtn: 'Tidak',
        users: {
          title: 'Pengguna',
          create: 'Buat Pengguna',
          enabled: 'Aktif',
          disabled: 'Nonaktif',
          deviceOwner: 'Pemilik perangkat',
          resetPassword: 'Atur Ulang Kata Sandi',
          delete: 'Hapus',
          deleteConfirm: 'Hapus pengguna ini dan cabut semua sesinya?',
          created: 'Pengguna dibuat',
          deleted: 'Pengguna dihapus',
          passwordUpdated: 'Kata sandi diperbarui',
          loadFailed: 'Gagal memuat pengguna',
          saveFailed: 'Gagal menyimpan pengguna',
          deleteFailed: 'Gagal menghapus pengguna'
        }
      },
      apiKeys: {
        mcpNote:
          'Kunci ini tidak berlaku untuk MCP, yang memiliki kuncinya sendiri di halaman MCP.',
        metricsUrl: 'URL metrik',
        monitoring: 'Pemantauan',
        monitoringDesc:
          'Prometheus membaca metrik dengan kunci API dari halaman ini, dikirim sebagai token Bearer. Semua peran dapat membacanya.',
        scrapeConfig: 'Konfigurasi scrape Prometheus',
        title: 'Kunci API',
        description:
          'Kunci bertindak sebagai pemiliknya, dengan peran pengguna tersebut. Kirimkan sebagai Authorization: Bearer <key> untuk metrik dan API, atau sebagai X-Auth-Token untuk Redfish.',
        name: 'Nama',
        namePlaceholder: 'Kegunaan kunci, misalnya prometheus',
        nameRequired: 'Beri nama kunci ini',
        nameTooLong: 'Nama maksimal 64 karakter',
        unnamed: '(tanpa nama)',
        create: 'Buat Kunci',
        created: 'Dibuat',
        owner: 'Pemilik',
        empty: 'Tidak ada kunci API',
        newKeyTitle: 'Kunci API baru Anda',
        newKeyWarning:
          'Salin kunci sekarang. Kunci tidak disimpan dan tidak dapat ditampilkan lagi. Jika hilang, cabut kunci tersebut dan buat yang baru.',
        copy: 'Salin',
        copied: 'Tersalin',
        copyFailed: 'Gagal menyalin. Salin secara manual.',
        done: 'Selesai',
        revoke: 'Cabut',
        revokeConfirmTitle: 'Cabut kunci API ini?',
        revokeConfirmDesc: 'Semua yang menggunakan "{{name}}" langsung berhenti berfungsi.',
        revoked: 'Kunci API dicabut',
        loadFailed: 'Gagal memuat kunci API',
        createFailed: 'Gagal membuat kunci API',
        revokeFailed: 'Gagal mencabut kunci API',
        cancelBtn: 'Batal'
      }
    },
    picoclaw: {
      title: 'PicoClaw Asisten',
      empty: 'Buka panel dan mulai tugas untuk memulai.',
      inputPlaceholder: 'Jelaskan apa yang Anda ingin PicoClaw lakukan',
      newConversation: 'Percakapan baru',
      processing: 'Memproses...',
      agent: {
        defaultTitle: 'Asisten umum',
        defaultDescription: 'Obrolan umum, pencarian, dan bantuan ruang kerja.',
        kvmTitle: 'Kontrol jarak jauh',
        kvmDescription: 'Operasikan host jarak jauh melalui IronKVM.',
        switched: 'Peran agen dialihkan',
        switchFailed: 'Gagal mengganti peran agen'
      },
      send: 'Kirim',
      cancel: 'Batalkan',
      status: {
        connecting: 'Menghubungkan ke gerbang...',
        connected: 'Sesi PicoClaw terhubung',
        disconnected: 'Sesi PicoClaw ditutup',
        stopped: 'Permintaan penghentian terkirim',
        runtimeStarted: 'Runtime PicoClaw dimulai',
        runtimeStartFailed: 'Gagal memulai Runtime PicoClaw',
        runtimeStopped: 'Runtime PicoClaw dihentikan',
        runtimeStopFailed: 'Gagal menghentikan Runtime PicoClaw',
        controlSwitchedToMCP: 'Kontrol dialihkan ke layanan MCP eksternal'
      },
      connection: {
        runtime: {
          checking: 'Memeriksa',
          restoring: 'Memulihkan PicoClaw',
          ready: 'Runtime siap',
          stopped: 'Runtime dihentikan',
          blockedByMCP: 'Kontrol MCP eksternal sedang aktif',
          readyBlockedByMCP:
            'Runtime berjalan, tetapi MCP eksternal saat ini mengendalikan input perangkat.',
          readyWithoutControl:
            'Runtime berjalan. Berikan kontrol perangkat ke PicoClaw sebelum menyambung ulang.',
          unavailable: 'Runtime tidak tersedia',
          configError: 'Kesalahan konfigurasi'
        },
        transport: {
          connecting: 'Menghubungkan',
          connected: 'Terhubung',
          disconnected: 'Terputus',
          reconnect: 'Sambungkan ulang',
          reconnectDescription: 'Sambungkan ulang ke sesi PicoClaw yang sedang berjalan.',
          reconnectBlocked: 'PicoClaw memerlukan kontrol perangkat sebelum menyambung ulang.'
        },
        run: {
          idle: 'Menganggur',
          busy: 'Sibuk'
        }
      },
      message: {
        toolAction: 'Aksi',
        observation: 'Pengamatan',
        screenshot: 'Tangkapan layar'
      },
      overlay: {
        locked: 'PicoClaw sedang mengendalikan perangkat. Input manual dijeda.'
      },
      control: {
        picoclaw: 'Kontrol perangkat: PicoClaw',
        picoclawDescription:
          'PicoClaw dapat mengirim input keyboard dan mouse. Input manual bisa dijeda.',
        mcp: 'Kontrol perangkat: MCP eksternal',
        mcpDescription:
          'MCP eksternal dapat menulis ke perangkat. PicoClaw tidak akan mengambil alih input.',
        off: 'Kontrol perangkat: nonaktif',
        offDescription:
          'AI tidak akan mengirim input keyboard atau mouse. Kontrol manual tetap tersedia.',
        transitioning: 'Kontrol perangkat: beralih',
        transitioningDescription: 'Kontrol perangkat sedang disinkronkan. Harap tunggu.',
        grant: 'Berikan kontrol',
        release: 'Lepaskan',
        releasing: 'Melepaskan...',
        switching: 'Beralih...',
        releasingLabel: 'Kontrol perangkat: melepaskan',
        releasingDescription:
          'Kontrol perangkat sedang dikembalikan. PicoClaw telah menghentikan penulisan yang berjalan.',
        granted: 'Kontrol PicoClaw diberikan',
        released: 'Kontrol PicoClaw dilepaskan',
        grantFailed: 'Gagal memberikan kontrol PicoClaw',
        releaseFailed: 'Gagal melepaskan kontrol PicoClaw',
        grantConfirmTitle: 'Alihkan kontrol perangkat ke PicoClaw?',
        grantConfirmDesc: 'Penulisan perangkat MCP eksternal akan dihentikan.'
      },
      install: {
        install: 'Instal PicoClaw',
        installing: 'Menginstal PicoClaw',
        success: 'PicoClaw berhasil diinstal',
        failed: 'Gagal menginstal PicoClaw',
        uninstalling: 'Mencopot pemasangan runtime...',
        uninstalled: 'Runtime berhasil di-uninstall.',
        uninstallFailed: 'Pencopotan pemasangan gagal.',
        requiredTitle: 'PicoClaw tidak diinstal',
        requiredDescription: 'Instal PicoClaw sebelum memulai runtime PicoClaw.',
        progressDescription: 'PicoClaw sedang diunduh dan diinstal.',
        stages: {
          preparing: 'Mempersiapkan',
          downloading: 'Mengunduh',
          extracting: 'Mengekstraksi',
          verifying: 'Memverifikasi',
          installing: 'Memasangkan',
          installed: 'Terpasang',
          install_timeout: 'Waktu Habis',
          install_failed: 'Gagal'
        }
      },
      model: {
        requiredTitle: 'Konfigurasi model diperlukan',
        requiredDescription: 'Konfigurasikan model PicoClaw sebelum menggunakan obrolan PicoClaw.',
        docsTitle: 'Panduan Konfigurasi',
        docsDesc: 'Model dan protokol yang didukung',
        menuLabel: 'Konfigurasi model',
        modelIdentifier: 'Pengenal Model',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Kunci API',
        apiKeyPlaceholder: 'Masukkan kunci API model',
        save: 'Simpan',
        saving: 'Menyimpan',
        saved: 'Konfigurasi model disimpan',
        saveFailed: 'Gagal menyimpan konfigurasi model',
        invalid: 'Pengidentifikasi model, API Base URL, dan kunci API wajib diisi'
      },
      uninstall: {
        menuLabel: 'Copot pemasangan',
        confirmTitle: 'Copot pemasangan PicoClaw',
        confirmContent:
          'Apakah Anda yakin ingin menghapus instalan PicoClaw? Ini akan menghapus semua file yang dapat dieksekusi dan konfigurasi.',
        confirmOk: 'Copot pemasangan',
        confirmCancel: 'Batalkan'
      },
      history: {
        title: 'Riwayat',
        loading: 'Memuat sesi...',
        emptyTitle: 'Belum ada riwayat',
        emptyDescription: 'Sesi PicoClaw sebelumnya akan muncul di sini.',
        loadFailed: 'Gagal memuat riwayat sesi',
        deleteFailed: 'Gagal menghapus sesi',
        deleteConfirmTitle: 'Hapus sesi',
        deleteConfirmContent: 'Apakah Anda yakin ingin menghapus "{{title}}"?',
        deleteConfirmOk: 'Hapus',
        deleteConfirmCancel: 'Batalkan',
        messageCount_other: '{{count}} pesan',
        messageCount: '{{count}} pesan'
      },
      config: {
        startRuntime: 'Mulai PicoClaw',
        stopRuntime: 'Hentikan PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Alihkan kontrol ke PicoClaw?',
        enableConfirmDesc: 'Memulai PicoClaw akan menonaktifkan layanan MCP eksternal.',
        enableConfirmOk: 'Mulai PicoClaw',
        enableConfirmCancel: 'Batal',
        title: 'Mulai PicoClaw',
        description: 'Mulai runtime untuk mulai menggunakan asisten PicoClaw.',
        switchFromMCP: 'Beralih ke PicoClaw dan mulai',
        takeoverAndStart: 'Ambil alih dan mulai'
      }
    },
    error: {
      title: 'Kami mengalami masalah',
      refresh: 'Segarkan',
      panel: 'Bagian halaman ini berhenti berfungsi',
      retry: 'Coba lagi'
    },
    fullscreen: {
      toggle: 'Beralih Layar Penuh'
    },
    input: {
      disconnected: 'Keyboard dan tetikus tidak terhubung',
      disconnectedTls:
        'Browser menolak koneksi aman yang membawa keyboard dan tetikus, dan melakukannya tanpa bertanya. Sertifikat yang dibuat perangkat ini belum dipercaya. Buka alamat ini di tab baru, terima sertifikatnya, lalu muat ulang. Memasang sertifikat adalah solusi yang andal.',
      disconnectedNever:
        'Koneksi yang membawa keyboard dan tetikus tidak dapat dibuka. Bagian lain halaman tetap berfungsi karena tidak menggunakannya. Pastikan tidak ada yang memblokirnya di antara Anda dan perangkat.',
      disconnectedDropped:
        'Koneksi yang membawa keyboard dan tetikus terputus dan belum pulih. Koneksi tersambung kembali sendiri setelah restart; jika tetap begini, muat ulang halaman.',
      hidDisabled: 'HID dinonaktifkan di perangkat ini (/boot/disable_hid).',
      keyFailed: 'Tombol tidak dapat dikirim.'
    },
    speaker: {
      title: 'Speaker',
      unmute: 'Bunyikan',
      mute: 'Bisukan',
      hostIdle: 'Host tidak mengirim audio',
      hostIdleHint: 'Putar sesuatu di host, atau pilih KVM sebagai output suaranya.'
    },
    upstream: {
      check: 'Periksa pembaruan',
      updateTo: 'Perbarui ke {{version}}',
      confirm: 'Perbarui {{name}} ke {{version}}?',
      confirmDesc:
        'Rilis baru diunduh dari GitHub dan diperiksa dengan checksum yang diterbitkannya. Jika ada yang gagal, versi saat ini tetap dipakai.',
      ok: 'Perbarui',
      upToDate: 'Sudah terbaru',
      builtIn: 'bawaan',
      checkFailed: 'Tidak dapat memeriksa pembaruan: {{error}}',
      unverifiable: 'Versi {{version}} tidak ditawarkan: {{reason}}',
      inUse: 'Tidak dapat memperbarui sekarang: {{reason}}',
      running: 'Memperbarui ke {{version}}...',
      done: '{{name}} diperbarui ke {{version}}',
      failed: 'Pembaruan terakhir gagal: {{error}}'
    },
    menu: {
      mediaAdd: 'Tambah image',
      mediaMoreOptions: 'Opsi lainnya',
      mediaSettings: 'Pengaturan media',
      collapse: 'Tutup Menu',
      expand: 'Perluas Menu',
      more: 'Lainnya',
      media: 'Media',
      tools: 'Alat',
      text: 'Teks',
      advanced: 'Lanjutan',
      mediaMounted: 'Terpasang',
      mediaLibrary: 'Pustaka',
      textToHost: 'Ke host',
      textFromHost: 'Dari host'
    },
    ion: {
      checking: 'Memeriksa memori video sebelum memulai streaming...',
      warn: 'Memori video hampir habis. Satu kali restart server akan menghabiskannya. Mulai ulang saat memungkinkan.',
      criticalTitle: 'Memori video tidak cukup untuk memulai streaming',
      criticalBody:
        'Memulai video akan menghabiskan memori yang dicadangkan dan menghentikan server. Semua fungsi lain tetap berjalan, termasuk kontrol daya dan mulai ulang. Hanya memulai ulang IronKVM yang dapat mengambil kembali memori ini.',
      criticalContinue: 'Tetap mulai video',
      criticalReboot: 'Mulai ulang IronKVM',
      criticalRebooting: 'Memulai ulang...'
    }
  }
};

export default id;
