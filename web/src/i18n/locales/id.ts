const id = {
  translation: {
    head: {
      desktop: 'Desktop jarak jauh',
      login: 'Masuk',
      changePassword: 'Ubah Sandi',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
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
          'To reset the passwords, pressing and holding the BOOT button on the NanoKVM for 10 seconds.',
        reset2: 'Untuk langkah-langkah rinci, lihat dokumen ini:',
        reset3: 'Akun web default:',
        reset4: 'Akun SSH default:',
        change1: 'Perhatikan bahwa tindakan ini akan mengubah kata sandi berikut:',
        change2: 'Kata sandi login web',
        change3: 'Kata sandi root sistem (kata sandi login SSH)',
        change4: 'Untuk mengatur ulang kata sandi, tekan dan tahan tombol BOOT pada NanoKVM.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Konfigurasi Wi-Fi untuk NanoKVM',
      success: 'Please check the network status of NanoKVM and visit the new IP address.',
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
      }
    },
    screen: {
      scale: 'Skala',
      title: 'Layar',
      video: 'Mode Video',
      videoDirectTips: 'Aktifkan HTTPS di "Pengaturan > Perangkat" untuk menggunakan mode ini',
      resolution: 'Resolusi',
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
      qualityLossless: 'Tanpa Kehilangan',
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
      }
    },
    keyboard: {
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
      absoluteStalled: 'Target mengabaikan tetikus absolut',
      absoluteStalledDesc:
        'Target berhenti menerima laporan tetikus absolut, sehingga gerakan penunjuk hilang. Keyboard tidak terpengaruh. Memulihkan USB biasanya mengatasinya; mode relatif menggunakan endpoint yang berbeda.',
      useRelative: 'Beralih ke mode relatif',
      direction: 'Arah roda gulir',
      scrollUp: 'Gulir ke atas',
      scrollDown: 'Gulir ke bawah',
      speed: 'Kecepatan roda gulir',
      fast: 'Cepat',
      slow: 'Lambat',
      requestPointer:
        'Menggunakan mode relatf. Silakan klik desktop untuk mendapatkan penunjuk tetikus.',
      resetHid: 'Setel ulang HID',
      hidOnly: {
        title: 'Mode hanya HID',
        desc: 'Jika mouse dan keyboard Anda berhenti merespons dan menyetel ulang HID tidak membantu, mungkin ada masalah kompatibilitas antara NanoKVM dan perangkat. Coba aktifkan mode HID-Only untuk kompatibilitas yang lebih baik.',
        tip1: 'Mengaktifkan mode HID-Hanya akan melepas U-disk virtual dan jaringan virtual',
        tip2: 'Dalam mode HID-Only, pemasangan gambar dinonaktifkan',
        rebuild: 'Mengganti mode akan membangun ulang koneksi USB. NanoKVM tidak dimulai ulang',
        enable: 'Aktifkan mode HID-Hanya',
        disable: 'Nonaktifkan mode HID-Hanya'
      }
    },
    image: {
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
      tips: {
        title: 'Cara mengunggah',
        usb1: 'Hubungkan NanoKVM ke komputer Anda melalui USB.',
        usb2: 'Pastikan disk virtual telah terpasang (Pengaturan - Disk Virtual).',
        usb3: 'Buka disk virtual di komputer Anda dan salin file gambar ke direktori root disk virtual.',
        scp1: 'Pastikan NanoKVM dan komputer Anda berada di jaringan lokal yang sama.',
        scp2: 'Buka terminal di komputer Anda dan gunakan perintah SCP untuk mengunggah file gambar ke direktori /data di NanoKVM.',
        scp3: 'Contoh: scp jalur-gambar-anda root@ip-nanokvm-anda:/data',
        tfCard: 'Kartu TF',
        tf1: 'Metode ini didukung di sistem linux',
        tf2: 'Dapatkan Kartu TF dari NanoKVM (untuk versi LENGKAP, bongkar casingnya terlebih dahulu).',
        tf3: 'Masukkan Kartu TF ke pembaca kartu dan hubungkan ke komputer Anda.',
        tf4: 'Salin berkas gambar ke direktori /data pada Kartu TF.',
        tf5: 'Masukkan Kartu TF ke dalam NanoKVM.'
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
      close: 'Tutup'
    },
    terminal: {
      title: 'Terminal',
      nanokvm: 'Terminal NanoKVM',
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
      title: 'Wake-on-LAN',
      sending: 'Kirim perintah...',
      sent: 'Perintah terkirim',
      input: 'Silahkan masukkan MAC',
      ok: 'Ok'
    },
    download: {
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
      cancelFailed: 'Gagal membatalkan unduhan'
    },
    power: {
      title: 'Daya',
      showConfirm: 'Konfirmasi',
      showConfirmTip: 'Pengoperasian listrik memerlukan konfirmasi tambahan',
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
      ledConnectedFailed: 'Gagal menyimpan pengaturan LED daya'
    },
    settings: {
      title: 'Pengaturan',
      mcp: {
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
        cancelBtn: 'Batal'
      },
      redfish: {
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
        actionReset: 'Reset',
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
        failed: 'Operasi watchdog gagal'
      },
      about: {
        title: 'Tentang NanoKVM',
        information: 'Informasi',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versi Aplikasi',
        applicationTip: 'Versi aplikasi web NanoKVM',
        image: 'Version Gambar',
        imageTip: 'Versi image sistem NanoKVM',
        kernel: 'Versi Kernel',
        kernelTip: 'Rilis kernel Linux yang sedang berjalan',
        deviceKey: 'Kunci Perangkat',
        videoMemory: 'Memori Video',
        videoMemoryTip:
          'Memori yang dicadangkan untuk penangkapan video. Memori ini tidak dibagi dengan bagian sistem lainnya.',
        videoMemoryGenerations_other:
          '{{count}} sesi NanoKVM sebelumnya masih menahan memori video',
        videoMemoryReboot: 'Mulai ulang untuk mengambilnya kembali.',
        community: 'Komunitas',
        hostname: 'Nama Host',
        hostnameUpdated: 'Nama host diperbarui. Nyalakan ulang untuk menerapkan.',
        ipType: {
          Wired: 'Berkabel',
          Wireless: 'Nirkabel',
          Other: 'Lainnya'
        }
      },
      appearance: {
        title: 'Tampilan',
        display: 'Layar',
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
          15: '15 sec',
          30: '30 sec',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 jam'
        },
        ssh: {
          description: 'Aktifkan akses jarak jauh SSH',
          tip: 'Tetapkan kata sandi yang kuat sebelum mengaktifkan (Akun - Ubah Kata Sandi)'
        },
        advanced: 'Pengaturan Lanjutan',
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
          tip: 'Mengaktifkan fitur ini dapat mempersingkat masa pakai kartu SD Anda!'
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
        autostart: {
          title: 'Pengaturan Skrip Mulai Otomatis',
          description: 'Mengelola skrip yang berjalan secara otomatis saat startup sistem',
          new: 'Baru',
          deleteConfirm: 'Apa kamu yakin menghapus data ini?',
          yes: 'Ya',
          no: 'Tidak',
          scriptName: 'Nama Skrip Mulai Otomatis',
          scriptContent: 'Konten Skrip Mulai Otomatis',
          settings: 'Pengaturan'
        },
        hidOnly: 'HID-Mode Hanya',
        hidOnlyDesc: 'Berhenti meniru perangkat virtual, hanya mempertahankan kontrol dasar HID',
        disk: 'Disk virtual',
        diskDesc: 'Mount virtual U-disk on the remote host',
        network: 'Jaringan virtual',
        networkDesc: 'Pasang kartu jaringan virtual pada host jarak jauh',
        usbNetwork: {
          description:
            'Tautan jaringan privat ke host jarak jauh melalui kabel USB. Host mendapat alamat tanpa gateway dan tanpa DNS, sehingga tidak dapat menjangkau LAN Anda melalui NanoKVM.',
          off: 'Mati',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (untuk host tanpa NCM)',
          rndis: 'RNDIS (tidak lagi ditawarkan)',
          rndisNote: 'Tautan ini memakai RNDIS, yang tidak lagi ditawarkan. Pilih NCM atau ECM.',
          subnet: 'Subnet',
          subnetDesc:
            'Jaringan IPv4 privat, /24 hingga /30. NanoKVM memakai alamat pertama, host memakai alamat kedua.',
          addresses: 'NanoKVM: {{board}}, host: {{host}}',
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
          'Menyediakan port serial USB untuk host jarak jauh, untuk masuk ke NanoKVM ini saat jaringan tidak dapat dijangkau',
        consoleTip:
          'Siapa pun yang mengendalikan host jarak jauh akan mendapatkan prompt login NanoKVM ini. Tetapkan kata sandi yang kuat sebelum mengaktifkan (Akun - Ubah Kata Sandi).',
        endpoints: {
          title: 'Endpoint USB',
          used: '{{used}} dari {{total}} terpakai',
          cost: 'memakai {{cost}}',
          needs: 'butuh {{cost}}',
          full: 'Endpoint USB tidak cukup. Nonaktifkan yang lain terlebih dahulu.',
          inactive:
            'Aktif, tetapi tidak berjalan: pengontrol USB kehabisan endpoint. Nonaktifkan perangkat lain dan perangkat ini akan langsung berjalan.',
          explain:
            'Pengontrol USB memiliki jumlah endpoint masuk yang tetap, dan inilah hitungannya. Jika perangkat yang diaktifkan melebihi kapasitas, keyboard dan tetikus dipertahankan dan sisanya dinonaktifkan.',
          error: 'Tidak dapat menjangkau perangkat. Coba lagi.',
          fitTogether: 'Yang muat bersamaan: {{sets}}'
        },
        reboot: 'Mulai ulang',
        rebootDesc: 'Apakah Anda yakin ingin me-reboot NanoKVM?',
        okBtn: 'Ya',
        cancelBtn: 'Tidak'
      },
      network: {
        title: 'Jaringan',
        wifi: {
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
            'Beralih kembali ke http. Muat ulang halaman ini jika tidak terbuka dengan sendirinya.'
        },
        ethernet: {
          title: 'Alamat IP',
          description: 'Atur cara NanoKVM memperoleh alamatnya di jaringan kabel',
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
          applyTitle: 'Ubah alamat NanoKVM?',
          applyWarning:
            'Koneksi ke halaman ini akan terputus. NanoKVM menerapkan alamat baru dan menunggu {{seconds}} detik sampai Anda menjangkaunya di sana. Menjangkaunya akan mempertahankan perubahan. Jika tidak ada yang menjangkaunya, NanoKVM mengembalikan pengaturan sebelumnya.',
          applyConfirm: 'Terapkan',
          applyCancel: 'Batal',
          applyFailed: 'Gagal menerapkan alamat',
          trialTitle: 'Menunggu konfirmasi',
          trialDhcp: 'NanoKVM sedang meminta alamat melalui DHCP.',
          trialStatic: 'NanoKVM sekarang berada di {{address}}.',
          trialInstruction:
            'Buka NanoKVM di alamat barunya dan masuk jika diminta. Menjangkaunya di sana akan mempertahankan perubahan. Jika tidak ada yang menjangkau NanoKVM dalam {{seconds}} detik, NanoKVM mengembalikan pengaturan sebelumnya.',
          trialOpen: 'Buka alamat baru',
          trialKeep: 'Pertahankan pengaturan ini',
          trialKept: 'Alamat baru tersimpan',
          trialKeepFailed: 'Gagal mempertahankan pengaturan',
          trialGone: 'Perubahan sudah dikembalikan. Coba lagi.',
          unsaved: 'Perubahan belum disimpan'
        },
        dns: {
          title: 'DNS',
          description: 'Konfigurasi server DNS untuk NanoKVM',
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
        enable: 'Aktifkan {{name}}',
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
          tip: 'Jika daemon kekurangan memori, coba aktifkan memori swap. Ini mengatur ukuran file swap menjadi 256MB secara default, yang dapat diubah di "Pengaturan > Perangkat".'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Harap segarkan dan coba lagi. Atau coba instal secara manual',
        download: 'Mengunduh',
        package: 'paket instalasi',
        unzip: 'dan unzip itu',
        upTailscale: 'Unggah tailscale ke direktori NanoKVM /usr/bin/',
        upTailscaled: 'Unggah tailscaled ke direktori NanoKVM /usr/sbin/',
        refresh: 'Segarkan halaman ini',
        notLogin:
          'Perangkat belum ditautkan. Silakan masuk dan tautkan perangkat ini ke akun Anda.',
        urlPeriod: 'Url ini berlaku selama 10 menit',
        login: 'Masuk',
        loginSuccess: 'Berhasil masuk',
        logout: 'Keluar',
        logoutDesc: 'Apakah Anda yakin ingin logout?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Perangkat ini belum bergabung ke jaringan NetBird. Bergabunglah dengan setup key, atau masuk dengan SSO.',
        setupKey: 'Setup key',
        setupKeyPlaceholder: 'Tempel setup key dari dasbor NetBird',
        join: 'Gabung',
        or: 'atau',
        sso: 'Masuk dengan SSO',
        urlPeriod: 'Url ini berlaku selama 10 menit',
        loginSuccess: 'Berhasil masuk',
        logout: 'Batalkan pendaftaran',
        logoutDesc:
          'Membatalkan pendaftaran akan menghapus peer ini dari akun NetBird Anda dan menghapus konfigurasinya di sini. Untuk bergabung lagi diperlukan setup key atau login SSO, dan peer mungkin mendapat IP baru. Lanjutkan?'
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
            'SHA-512 hanya memeriksa bahwa paket cocok dengan manifes yang disediakan oleh server ini. Pemeriksaan ini tidak membuktikan bahwa paket tersebut merupakan rilis resmi NanoKVM. Server yang bermasalah atau berbahaya dapat membuat perangkat tidak dapat digunakan, menyebabkan kehilangan data, atau membahayakan sistem.',
          confirm: 'Tetap Gunakan',
          useSipeed: 'Gunakan server resmi Sipeed',
          previewDisabled:
            'Pembaruan Pratinjau tidak tersedia saat server pembaruan kustom diaktifkan.'
        },
        offline: {
          title: 'Pembaruan Offline',
          desc: 'Perbarui melalui paket instalasi lokal',
          upload: 'Mengunggah',
          checksumPlaceholder: 'Checksum SHA-256 (opsional)',
          invalidChecksum: 'Checksum SHA-256 harus berisi 64 karakter heksadesimal.',
          checksumMismatch: 'Verifikasi SHA-256 gagal. Paket mungkin rusak.',
          invalidName: 'Format nama file tidak valid. Silakan unduh dari rilis GitHub.',
          updateFailed: 'Gagal memperbarui, tolong coba lagi.'
        }
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
        kvmDescription: 'Operasikan host jarak jauh melalui NanoKVM.',
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
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime siap',
          stopped: 'Runtime dihentikan',
          blockedByMCP: 'Kontrol MCP eksternal sedang aktif',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime tidak tersedia',
          configError: 'Kesalahan konfigurasi'
        },
        transport: {
          connecting: 'Menghubungkan',
          connected: 'Terhubung',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
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
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Kontrol perangkat: MCP eksternal',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Kontrol perangkat: nonaktif',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Berikan kontrol',
        release: 'Lepaskan',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
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
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
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
    speaker: { title: 'Speaker', unmute: 'Bunyikan', mute: 'Bisukan' },
    menu: {
      collapse: 'Tutup Menu',
      expand: 'Perluas Menu'
    },
    ion: {
      checking: 'Memeriksa memori video sebelum memulai streaming...',
      warn: 'Memori video hampir habis. Satu kali restart server akan menghabiskannya. Mulai ulang saat memungkinkan.',
      criticalTitle: 'Memori video tidak cukup untuk memulai streaming',
      criticalBody:
        'Memulai video akan menghabiskan memori yang dicadangkan dan menghentikan server. Semua fungsi lain tetap berjalan, termasuk kontrol daya dan mulai ulang. Hanya memulai ulang NanoKVM yang dapat mengambil kembali memori ini.',
      criticalContinue: 'Tetap mulai video',
      criticalReboot: 'Mulai ulang NanoKVM',
      criticalRebooting: 'Memulai ulang...'
    }
  }
};

export default id;
