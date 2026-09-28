const pt_br = {
  translation: {
    head: {
      desktop: 'Área de Trabalho Remota',
      login: 'Login',
      changePassword: 'Mudar Senha',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      cookieRejected:
        'O navegador se recusou a armazenar a sessão. Um cookie deixado por uma sessão HTTPS anterior não pode ser substituído por http simples. Limpe os cookies deste endereço, ou abra uma janela anônima, e entre novamente.',
      login: 'Login',
      placeholderUsername: 'Nome de usuário',
      placeholderPassword: 'Senha',
      placeholderCurrentPassword: 'Senha atual',
      placeholderPassword2: 'Por favor, digite a senha novamente',
      noEmptyUsername: 'Nome de usuário é obrigatório',
      noEmptyPassword: 'Senha é obrigatória',
      passwordLength: 'A senha deve ter entre 8 e 72 caracteres',
      noAccount:
        'Falha ao obter informações do usuário, por favor atualize a página ou redefina a senha',
      invalidUser: 'Nome de usuário ou senha inválidos',
      locked: 'Muitos logins, tente novamente mais tarde',
      globalLocked: 'Sistema sob proteção, tente novamente mais tarde',
      error: 'Erro inesperado',
      invalidCurrentPassword: 'A senha atual está incorreta',
      changePassword: 'Mudar Senha',
      changePasswordDesc: 'Para a segurança do seu dispositivo, por favor, mude a senha!',
      differentPassword: 'Senhas não conferem',
      illegalUsername: 'Nome de usuário contém caracteres inválidos',
      illegalPassword: 'Senha contém caracteres inválidos',
      forgetPassword: 'Esqueci a senha',
      ok: 'Ok',
      cancel: 'Cancelar',
      loginButtonText: 'Login',
      tips: {
        reset1:
          'Para redefinir as senhas, pressione e segure o botão BOOT no NanoKVM por 10 segundos.',
        reset2: 'Para etapas detalhadas, por favor, consulte este documento:',
        reset3: 'Conta padrão da Web:',
        reset4: 'Conta padrão SSH:',
        change1: 'Por favor, note que esta ação irá alterar as seguintes senhas:',
        change2: 'Senha de login da Web',
        change3: 'Senha root do sistema (senha de login SSH)',
        change4: 'Para redefinir as senhas, pressione e segure o botão BOOT no NanoKVM.'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Configurar Wi-Fi para o NanoKVM',
      success: 'Por favor, verifique o status da rede do NanoKVM e visite o novo endereço IP.',
      failed: 'Operação falhou, por favor, tente novamente.',
      invalidMode:
        'O modo atual não suporta configuração de rede. Vá para o seu dispositivo e ative o modo de configuração Wi-Fi.',
      confirmBtn: 'Ok',
      finishBtn: 'Finalizado',
      ap: {
        authTitle: 'Autenticação necessária',
        authDescription: 'Por favor, digite a senha AP para continuar',
        authFailed: 'Senha AP inválida',
        passPlaceholder: 'AP senha',
        verifyBtn: 'Verificar'
      },
      ssidRequired: 'Digite o nome da rede, até 32 caracteres',
      passwordLength: 'A senha tem de 8 a 63 caracteres. Deixe vazia para uma rede aberta.',
      passwordOptional: 'Senha (vazia para uma rede aberta)',
      lost:
        'A placa parou de responder. Ela pode ter entrado na rede e fechado seu hotspot de configuração. Se o hotspot voltar, a conexão falhou: conecte-se a ele de novo e tente outra vez.',
      done:
        'Configuração concluída. Reconecte este dispositivo à sua rede habitual e abra a placa em seu novo endereço.'
    },
    screen: {
      scale: 'Escala',
      title: 'Tela',
      video: 'Modo de Vídeo',
      videoDirectTips: 'Ative HTTPS em "Configurações > Dispositivo" para usar este modo',
      resolution: 'Resolução',
      ocr: {
        title: 'Ler texto (OCR)',
        tips: 'O texto é reconhecido neste navegador. Você pode corrigi-lo antes de copiar.',
        hint: 'Arraste sobre o texto a ser lido. Pressione Esc para cancelar.',
        noPicture: 'Aguarde o vídeo e depois arraste sobre o texto a ser lido.',
        cancel: 'Cancelar',
        language: 'Idioma',
        languages: {
          eng: 'Inglês'
        },
        preview: 'Área selecionada',
        capturing: 'Capturando a tela...',
        loading: 'Carregando o reconhecimento de texto...',
        recognizing: 'Lendo o texto...',
        noText: 'Nenhum texto foi encontrado na área selecionada.',
        copy: 'Copiar',
        copied: 'Copiado para a área de transferência',
        copyFailed: 'Não foi possível copiar para a área de transferência',
        selectAgain: 'Selecionar novamente',
        unsupported:
          'Este navegador não consegue executar o reconhecimento de texto. Ele precisa de WebAssembly SIMD, que os navegadores atuais têm.',
        captureFailed: 'Não foi possível capturar a tela.',
        outside: 'A área selecionada está fora da imagem.',
        recognizeFailed: 'O reconhecimento de texto falhou.'
      },
      controlRegion: {
        title: 'Calibração do mouse',
        description:
          'Use esta configuração quando o dispositivo controlado usar uma resolução diferente de 16:9 e o cursor estiver desalinhado horizontal ou verticalmente.',
        off: 'Desativado',
        auto: 'Automático',
        autoWarning:
          'A calibração pode falhar se o aplicativo do usuário tiver um fundo totalmente preto.',
        manual: 'Manual',
        selectedResolution: 'Resolução da área selecionada',
        unused: 'Não utilizado',
        originalResolution: 'Resolução original',
        selectResolution: 'Selecionar resolução original',
        addResolution: 'Adicionar resolução personalizada',
        add: 'Adicionar',
        duplicateResolution: 'Esta resolução já existe.',
        width: 'Largura',
        height: 'Altura',
        apply: 'Calcular e aplicar',
        invalidResolution: 'Insira uma resolução original válida quando o vídeo estiver pronto.',
        select: 'Selecionar área',
        clear: 'Restaurar detecção automática',
        saveFailed: 'Falha ao salvar a área de entrada.',
        tooSmall: 'A área selecionada é muito pequena.',
        previewUnavailable: 'Pré-visualização indisponível',
        clearConfirm: 'Restaurar a detecção automática de bordas pretas?',
        dragHint: 'Arraste para selecionar a área da área de trabalho remota',
        finish: 'Concluir',
        confirm: 'Confirmar',
        cancel: 'Cancelar'
      },
      auto: 'Automático',
      autoTips:
        'Rasgos na tela ou desvio do mouse podem ocorrer em resoluções específicas. Considere ajustar a resolução do host remoto ou desativar o modo automático.',
      fps: 'FPS',
      customizeFps: 'Personalizar',
      quality: 'Qualidade',
      qualityLossless: 'Sem perdas',
      qualityHigh: 'Alta',
      qualityMedium: 'Média',
      qualityLow: 'Baixa',
      frameDetect: 'Detecção de Quadros',
      frameDetectTip:
        'Calcular a diferença entre os quadros. Parar a transmissão de vídeo quando nenhuma alteração for detectada na tela do host remoto.',
      resetHdmi: 'Redefinir HDMI',
      mixedH264: {
        title: 'Conflito de transmissão H.264',
        description:
          'H.264 Direct e H.264 WebRTC estão sendo usados ao mesmo tempo. Isso pode causar rasgos na tela ou vídeo corrompido. Use apenas um modo H.264.'
      },
      webrtcConnectionFailed: {
        title: 'Falha na conexão WebRTC',
        description: 'Verifique a conexão de rede ou alterne o modo de vídeo.'
      },
      captureStatus: {
        hdmiError: 'Erro na imagem HDMI',
        unsupportedResolution: 'A resolução atual não é compatível',
        retrieving: 'Obtendo tela...',
        changingResolution: 'Alterando resolução...',
        updateFailed: 'A tela não pode ser atualizada agora',
        videoError: 'Erro na exibição de vídeo',
        noHdmi: 'Nenhum sinal HDMI detectado',
        unavailable: 'A tela não pode ser exibida agora'
      }
    },
    keyboard: {
      title: 'Teclado',
      paste: 'Colar',
      tips: 'Digita o texto no host como pressionamentos de teclas. Escolha o layout de teclado que o host usa.',
      placeholder: 'Por favor, digite',
      submit: 'Enviar',
      virtual: 'Teclado',
      readClipboard: 'Ler da área de transferência',
      clipboardPermissionDenied:
        'Permissão da área de transferência negada. Permita o acesso à área de transferência no seu navegador.',
      clipboardReadError: 'Falha ao ler a área de transferência',
      mediaKeys: {
        title: 'Teclas de mídia',
        mute: 'Mudo',
        volumeDown: 'Diminuir volume',
        volumeUp: 'Aumentar volume',
        previous: 'Faixa anterior',
        playPause: 'Reproduzir ou pausar',
        next: 'Próxima faixa',
        stop: 'Parar'
      },
      pasting: {
        layout: 'Layout de teclado do host',
        layouts: {
          us: 'Inglês (EUA)',
          uk: 'Inglês (Reino Unido)',
          de: 'Alemão',
          fr: 'Francês',
          es: 'Espanhol',
          it: 'Italiano',
          ptBr: 'Português (Brasil)',
          se: 'Sueco / finlandês',
          ru: 'Russo',
          ja: 'Japonês',
          ko: 'Coreano'
        },
        speed: 'Velocidade de digitação',
        speeds: {
          fast: 'Rápida',
          normal: 'Normal',
          slow: 'Lenta'
        },
        estimate: 'Tempo de digitação: cerca de {{duration}}',
        untypeable: 'Caracteres que este layout não consegue digitar: {{count}}',
        untypeableAt: 'linha {{line}}, coluna {{column}}',
        skipUntypeable: 'Digitar o resto',
        shortcut: '{{shortcut}} digita a área de transferência no host na hora.',
        clipboardUnavailable:
          'O navegador só deixa uma página ler a área de transferência via HTTPS. Cole o texto na caixa com Ctrl+V.',
        clipboardEmpty: 'A área de transferência não contém texto.',
        tooLong: 'O texto é longo demais. O limite é de {{max}} caracteres.',
        inProgress: 'Já há uma colagem sendo digitada.',
        typing: 'Digitando no host',
        done: 'Texto digitado',
        canceled: 'Colagem cancelada',
        failed: 'A colagem falhou',
        cancel: 'Cancelar',
        controlBusy: 'Outro controlador está usando o teclado.',
        hidError: 'Não foi possível enviar os pressionamentos de teclas ao host.'
      },
      shortcut: {
        title: 'Atalhos',
        custom: 'Personalizado',
        capture: 'Clique aqui para capturar o atalho',
        clear: 'Limpar',
        save: 'Salvar',
        captureTips:
          'Capturar teclas do sistema (como a tecla Windows) requer permissão de tela cheia.',
        enterFullScreen: 'Alternar modo de tela cheia.'
      },
      leaderKey: {
        title: 'Tecla Leader',
        desc: 'Ignore as restrições do navegador e envie atalhos do sistema diretamente para o host remoto.',
        howToUse: 'Como usar',
        simultaneous: {
          title: 'Modo Simultâneo',
          desc1: 'Pressione e segure a tecla Leader e depois pressione o atalho.',
          desc2: 'Intuitivo, mas pode entrar em conflito com atalhos do sistema.'
        },
        sequential: {
          title: 'Modo Sequencial',
          desc1:
            'Pressione a tecla Leader → pressione o atalho em sequência → pressione a tecla Leader novamente.',
          desc2: 'Requer mais etapas, mas evita completamente conflitos de sistema.'
        },
        enable: 'Habilitar tecla Leader',
        tip: 'Quando atribuída como tecla Leader, esta tecla funciona apenas como gatilho de atalho e perde seu comportamento padrão.',
        placeholder: 'Pressione a tecla Leader',
        shiftRight: 'Shift direito',
        ctrlRight: 'Ctrl direito',
        metaRight: 'Win direito',
        submit: 'Enviar',
        recorder: {
          rec: 'REC',
          activate: 'Ativar teclas',
          input: 'Por favor, pressione o atalho...'
        }
      }
    },
    mouse: {
      title: 'Mouse',
      cursor: 'Estilo do cursor',
      default: 'Cursor padrão',
      pointer: 'Cursor de ponteiro',
      cell: 'Cursor de célula',
      text: 'Cursor de texto',
      grab: 'Cursor de arrastar',
      hide: 'Ocultar cursor',
      mode: 'Modo do mouse',
      absolute: 'Modo absoluto',
      relative: 'Modo relativo',
      absoluteShort: 'Absoluto',
      relativeShort: 'Relativo',
      touch: 'Modo toque',
      touchShort: 'Toque',
      absoluteStalled: 'O alvo está ignorando o mouse absoluto',
      absoluteStalledDesc:
        'O alvo parou de receber os relatórios do mouse absoluto, então os movimentos do ponteiro são perdidos. O teclado não é afetado. Recuperar o USB costuma resolver; o modo relativo usa outro endpoint.',
      useRelative: 'Mudar para modo relativo',
      direction: 'Direção da roda de rolagem',
      scrollUp: 'Role para cima',
      scrollDown: 'Role para baixo',
      speed: 'Velocidade da roda de rolagem',
      fast: 'Rápido',
      slow: 'Lento',
      requestPointer:
        'Usando modo relativo. Por favor, clique na área de trabalho para obter o ponteiro do mouse.',
      resetHid: 'Redefinir HID',
      hidOnly: {
        title: 'Modo somente HID',
        desc: 'Se o seu mouse e teclado pararem de responder e a redefinição de HID não ajudar, pode ser um problema de compatibilidade entre o NanoKVM e o dispositivo. Tente habilitar o modo Somente-HID para melhor compatibilidade.',
        tip1: 'Habilitar o modo Somente-HID irá desmontar o U-disk virtual e a rede virtual',
        tip2: 'No modo Somente-HID, a montagem de imagem está desativada',
        rebuild: 'Trocar de modo reconstrói a conexão USB. O NanoKVM não reinicia',
        enable: 'Habilitar modo Somente-HID',
        disable: 'Desabilitar modo Somente-HID'
      }
    },
    image: {
      title: 'Imagens',
      loading: 'Carregando...',
      empty: 'Nada Encontrado',
      mountMode: 'Modo de montagem',
      mountFailed: 'Falha na Montagem',
      mountDesc:
        'Em alguns sistemas, é necessário ejetar o disco virtual no host remoto antes de montar a imagem.',
      unmountFailed: 'Falha na desmontagem',
      unmountDesc:
        'Em alguns sistemas, é necessário ejetar manualmente do host remoto antes de desmontar a imagem.',
      refresh: 'Atualizar a lista de imagens',
      disk: 'Disco',
      cdrom: 'CD',
      driveEmpty: 'Vazio',
      eject: 'Ejetar',
      readOnly: 'Somente leitura',
      readOnlyTip: 'Vale para a próxima imagem inserida no disco.',
      noDrives: 'Nenhuma unidade virtual. Ative o disco virtual em Configurações.',
      insertFailed: 'Falha ao inserir',
      ejectFailed: 'Falha ao ejetar',
      insertInto: 'Inserir em {{drive}}. Clique para alterar.',
      loadedIn: 'Na unidade {{drive}}',
      attention: 'Atenção',
      deleteConfirm: 'Tem certeza que deseja excluir esta imagem?',
      okBtn: 'Sim',
      cancelBtn: 'Não',
      deleteFailed: 'Falha ao excluir',
      ventoy: {
        statusNoKernel: 'Sem suporte neste firmware',
        statusNotInstalled: 'Não instalado',
        statusReady: 'Pronto',
        statusSelected: 'Imagens selecionadas: {{count}}',
        statusInDrive: 'Na unidade de disco, {{size}}',
        noKernel:
          'O kernel deste firmware não tem suporte a device-mapper, então o Ventoy não pode ser usado até que uma imagem com esse suporte seja instalada.',
        installDesc: 'Inicie o host a partir de várias imagens em um único disco, sem copiá-las.',
        install: 'Instalar',
        installing: 'Baixando o Ventoy, cerca de 20 MB. Isso pode levar alguns minutos.',
        needsData: 'O Ventoy precisa de uma imagem do IronKVM com a partição /data montada.',
        uninstall: 'Desinstalar',
        uninstallConfirm: 'Remover os arquivos do Ventoy?',
        noImages: 'Nenhuma imagem para colocar no disco Ventoy.',
        onDisk: 'No disco Ventoy',
        missing: 'Ausente: {{file}}',
        remove: 'Tirar do disco Ventoy',
        setHint:
          'O conjunto de imagens só muda enquanto o disco Ventoy não está em nenhuma unidade.',
        useAsDisk: 'Usar como disco virtual',
        failed: 'Falha na solicitação do Ventoy',
        secureBoot:
          'Com o Secure Boot ativado, o host precisa registrar a chave do Ventoy no MokManager uma vez. O arquivo de chave ENROLL_THIS_KEY_IN_MOKMANAGER.cer está na partição VTOYEFI.',
        readOnly:
          'O host vê o disco como somente leitura, então a persistência do Ventoy e o ventoy.json na unidade não funcionam.'
      },
      tips: {
        title: 'Como fazer upload',
        usb1: 'Conecte o NanoKVM ao seu computador via USB.',
        usb2: 'Certifique-se de que o disco virtual está montado (Configurações - Disco Virtual).',
        usb3: 'Abra o disco virtual no seu computador e copie o arquivo de imagem para o diretório raiz do disco virtual.',
        scp1: 'Certifique-se de que o NanoKVM e seu computador estão na mesma rede local.',
        scp2: 'Abra um terminal no seu computador e use o comando SCP para fazer upload do arquivo de imagem para o diretório /data no NanoKVM.',
        scp3: 'Exemplo: scp seu-caminho-da-imagem root@seu-ip-nanokvm:/data',
        tfCard: 'Cartão TF',
        tf1: 'Este método é suportado em sistemas Linux',
        tf2: 'Remova o cartão TF do NanoKVM (para a versão FULL, desmonte a caixa primeiro).',
        tf3: 'Insira o cartão TF em um leitor de cartão e conecte-o ao seu computador.',
        tf4: 'Copie o arquivo de imagem para o diretório /data no cartão TF.',
        tf5: 'Insira o cartão TF no NanoKVM.'
      }
    },
    script: {
      title: 'Scripts',
      upload: 'Upload',
      run: 'Executar',
      runBackground: 'Executar em segundo plano',
      runFailed: 'Falha na execução',
      attention: 'Atenção',
      delDesc: 'Tem certeza de que deseja excluir este arquivo?',
      confirm: 'Sim',
      cancel: 'Não',
      delete: 'Excluir',
      close: 'Fechar'
    },
    terminal: {
      title: 'Terminal',
      nanokvm: 'Terminal NanoKVM',
      serial: 'Terminal de Porta Serial',
      serialPort: 'Porta Serial',
      serialPortPlaceholder: 'Por favor, digite a porta serial',
      baudrate: 'Taxa de transmissão',
      parity: 'Paridade',
      parityNone: 'Nenhum',
      parityEven: 'Par',
      parityOdd: 'Ímpar',
      flowControl: 'Controle de fluxo',
      flowControlNone: 'Nenhum',
      flowControlSoft: 'Software',
      flowControlHard: 'Hardware',
      dataBits: 'Bits de dados',
      stopBits: 'Bits de parada',
      confirm: 'Ok'
    },
    wol: {
      title: 'Wake-on-LAN',
      sending: 'Enviando comando...',
      sent: 'Comando enviado',
      input: 'Por favor, digite o MAC',
      ok: 'Ok'
    },
    download: {
      title: 'Baixador de Imagens',
      input: 'Por favor, digite uma URL de imagem remota',
      ok: 'Ok',
      disabled: 'A partição /data é RO, então não podemos baixar a imagem',
      uploadbox: 'Solte o arquivo aqui ou clique para selecionar',
      inputfile: 'Por favor insira o arquivo de imagem',
      NoISO: 'Sem ISO',
      sha256: 'SHA-256 (opcional)',
      sha256Placeholder: 'Digite um checksum SHA-256 de 64 caracteres',
      invalidSHA256: 'SHA-256 deve ser uma sequência hexadecimal de 64 caracteres',
      failed: 'Falha no download',
      success: 'Download concluído',
      checksumFailed: 'Falha no download: a verificação SHA-256 falhou',
      cancel: 'Cancelar',
      cancelFailed: 'Falha ao cancelar o download',
      bootMenu: 'Menu de boot (netboot.xyz)',
      bootMenuDesc: 'Baixar a ISO do netboot.xyz, com checksum conferido, para o CD virtual'
    },
    power: {
      title: 'Energia',
      showConfirm: 'Confirmação',
      showConfirmTip: 'Perguntar antes de um toque curto. Reset e toque longo sempre perguntam.',
      reset: 'Redefinir',
      power: 'Energia',
      powerShort: 'Energia (clique curto)',
      powerLong: 'Energia (clique longo)',
      resetConfirm: 'Prosseguir com a operação de redefinição?',
      powerConfirm: 'Prosseguir com a operação de energia?',
      okBtn: 'Sim',
      cancelBtn: 'Não',
      hostOs: 'SO do host',
      hostOsTip: 'Enviadas como teclas USB. O host decide o que elas fazem.',
      sleep: 'Suspender',
      wake: 'Despertar',
      wakeKey: 'Despertar com Shift',
      powerDown: 'Desligar',
      sleepConfirm: 'Suspender o host?',
      powerDownConfirm: 'Enviar a tecla de desligar ao host?',
      wakeTip:
        'Um host suspenso costuma ignorar Despertar vindo do dispositivo que o suspendeu. Despertar com Shift pressiona uma tecla do teclado, que mais hosts aceitam.',
      led: 'LED de energia',
      ledOn: 'Aceso',
      ledOff: 'Apagado',
      ledUnknown: 'Desconhecido',
      ledConnected: 'LED de energia conectado',
      ledConnectedTip:
        'Ative somente se o conector do LED de energia do host estiver ligado à placa. Sem ele, o estado de energia é desconhecido.',
      ledConnectedFailed: 'Falha ao salvar a configuração do LED de energia',
      powerLongConfirm:
        'Segurar o botão de energia por {{seconds}} s? Isso corta a energia sem desligar o sistema.',
      done: 'Botão pressionado',
      failed: 'Falha ao pressionar o botão'
    },
    settings: {
      title: 'Configurações',
      mcp: {
        title: 'Serviço MCP',
        service: 'Controle remoto MCP',
        serviceDesc:
          'Permitir que clientes MCP confiáveis controlem o teclado e o mouse e capturem imagens da tela',
        securityWarning:
          'Qualquer pessoa com esta chave de API pode controlar o host remoto e visualizar sua tela. Use HTTPS e habilite o serviço somente em redes confiáveis.',
        endpoint: 'Endpoint',
        apiKey: 'Chave de API',
        regenerateConfirmTitle: 'Gerar novamente a chave de API MCP?',
        regenerateConfirmDesc: 'A chave atual deixará de funcionar imediatamente.',
        enableConfirmTitle: 'Habilitar o controle MCP externo?',
        enableConfirmDesc:
          'Habilitar o MCP interromperá o PicoClaw e fechará todas as sessões ativas do PicoClaw.',
        failed: 'Falha na operação MCP',
        copyFailed: 'Falha ao copiar. Copie manualmente.',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar'
      },
      redfish: {
        title: 'Redfish',
        service: 'Serviço Redfish',
        serviceDesc:
          'A API Redfish da DMTF, para controle de energia, mídia virtual e status a partir de ferramentas como redfishtool e Ansible. Desativá-la encerra todas as sessões Redfish.',
        endpoint: 'Raiz do serviço',
        httpsOn: 'A placa serve HTTPS, que a maioria das ferramentas Redfish exige.',
        httpsOff:
          'A placa serve HTTP simples. A maioria das ferramentas Redfish exige HTTPS: ative-o em "Configurações > Rede".',
        credentials:
          'O Redfish aceita as contas do KVM, com autenticação Basic ou uma sessão Redfish, e chaves de API enviadas como X-Auth-Token. As chaves de API são gerenciadas na página Chaves de API.',
        powerActions: 'Ações de energia',
        powerActionsDesc:
          'Os tipos de reset oferecidos agora. On, ForceOff e GracefulShutdown precisam do estado de energia, então só são oferecidos quando "LED de energia conectado" está ativado no menu de energia.',
        sessions: 'Sessões',
        noSessions: 'Nenhuma sessão Redfish aberta',
        created: 'Criada',
        lastUsed: 'Último uso',
        refresh: 'Atualizar',
        end: 'Encerrar',
        endConfirmTitle: 'Encerrar esta sessão Redfish?',
        endConfirmDesc:
          'O token dela para de funcionar imediatamente. O cliente precisa fazer login novamente.',
        failed: 'Falha na operação Redfish',
        copyFailed: 'Falha ao copiar. Copie manualmente.',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar'
      },
      ipmi: {
        title: 'IPMI',
        warning:
          'A autenticação IPMI é fraca por design. Qualquer pessoa que alcance a placa e conheça um nome de usuário pode obter um hash da senha IPMI desse usuário e tentar quebrá-lo offline. Use senhas geradas, ative o IPMI apenas em uma rede confiável e prefira Redfish sobre HTTPS quando a ferramenta suportar.',
        service: 'IPMI sobre LAN',
        serviceDesc:
          'IPMI 2.0 (RMCP+, ipmitool lanplus) na porta UDP 623, para a energia e o estado do host. IPMI 1.5 e o cipher suite 0 são recusados. Desativá-lo encerra todas as sessões IPMI.',
        example: 'Exemplo',
        copyFailed: 'Falha ao copiar. Copie manualmente.',
        ledOn: 'Estado de energia, on, off, soft, cycle e reset estão disponíveis.',
        ledOff:
          '"LED de energia conectado" está desligado no menu de energia, então o estado de energia é desconhecido. Só "power reset" funciona: status, on, off, soft e cycle são recusados.',
        accounts: 'Contas',
        accountsDesc:
          'O IPMI faz login com as contas do KVM, cada uma com sua própria senha IPMI, separada da senha web. Administradores recebem ADMINISTRATOR. Usuários recebem USER: podem ler o estado de energia com "-L USER", mas não alterá-lo.',
        passwordSet: 'Senha IPMI definida',
        passwordNotSet: 'Sem senha IPMI: não pode fazer login via IPMI',
        nameTooLong: 'O nome tem mais de 16 caracteres, o que o IPMI não permite',
        accountDisabled: 'A conta está desativada',
        setPassword: 'Definir senha',
        changePassword: 'Alterar senha',
        remove: 'Remover',
        removeConfirmTitle: 'Remover a senha IPMI de {{user}}?',
        removeConfirmDesc:
          'A conta não poderá mais fazer login via IPMI, e suas sessões IPMI serão encerradas.',
        passwordTitle: 'Senha IPMI de {{user}}',
        passwordDesc:
          'De 12 a 20 caracteres ASCII imprimíveis, diferente da senha web. O IPMI exige que a placa guarde a senha de forma que possa lê-la de volta, então use uma que não seja usada em nenhum outro lugar. Copie-a antes de salvar: ela não será mostrada novamente.',
        passwordPlaceholder: 'Senha IPMI',
        generate: 'Gerar',
        copy: 'Copiar',
        save: 'Salvar',
        passwordLength: 'Use de 12 a 20 caracteres.',
        passwordChars: 'Use apenas caracteres ASCII imprimíveis.',
        saved: 'Senha IPMI salva',
        failed: 'A operação IPMI falhou',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar'
      },
      vnc: {
        title: 'VNC',
        service: 'Servidor VNC',
        serviceDesc:
          'Permite que um cliente VNC, como TigerVNC ou Remmina, veja e controle o host. O cliente precisa suportar a codificação Tight. Uma sessão por vez.',
        credentials:
          'Entre com uma conta KVM. A conexão é criptografada com o certificado TLS da placa (VeNCrypt X509Plain).',
        port: 'Porta',
        portDesc: 'A porta TCP em que o servidor escuta.',
        maxFps: 'Limite de quadros',
        maxFpsDesc: 'O máximo de quadros por segundo enviados a um cliente.',
        vncAuth: 'Autenticação VNC simples',
        vncAuthDesc:
          'Para clientes sem VeNCrypt. Ela verifica uma senha VNC separada em vez de uma conta.',
        vncAuthWarning:
          'A autenticação VNC simples não criptografa a conexão. Qualquer pessoa no caminho da rede pode ver a tela e as teclas digitadas. Use-a apenas em uma rede confiável.',
        password: 'Senha VNC',
        passwordSet: 'Uma senha está definida. Digite uma nova para alterá-la.',
        passwordInvalid: 'A senha VNC deve ter de 6 a 8 caracteres.',
        save: 'Salvar',
        saved: 'Configurações salvas',
        state: 'Estado',
        listening: 'Escutando na porta {{port}}',
        notListening: 'Não está escutando',
        noSession: 'Nenhuma sessão aberta',
        client: 'Cliente',
        user: 'Usuário',
        method: 'Autenticação',
        methodVencrypt: 'Conta sobre TLS',
        methodVnc: 'Senha VNC',
        since: 'Conectado desde',
        resolution: 'Resolução',
        framesSent: 'Quadros enviados',
        lastError: 'A última sessão terminou: {{error}}',
        refresh: 'Atualizar',
        disconnect: 'Desconectar',
        disconnectConfirmTitle: 'Encerrar a sessão VNC?',
        disconnectConfirmDesc:
          'O cliente é desconectado imediatamente, e todas as teclas e botões pressionados são soltos.',
        failed: 'Falha na operação VNC',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar'
      },
      watchdog: {
        title: 'Watchdog',
        service: 'Watchdog do host',
        serviceDesc:
          'Se o host deveria estar ligado e a imagem dele não muda, ou não há sinal HDMI, durante o tempo limite, a placa pressiona reset ou desliga e religa o host.',
        stillWarning:
          'Um host cuja tela entra em repouso, ou cuja imagem fica parada enquanto trabalha, parece travado. Desative o repouso da tela no host ou defina um endereço de ping.',
        ledHint:
          '"LED de energia conectado" está desativado no menu de energia. O watchdog não vê quando o host está desligado, então o trata como sempre ligado.',
        timeout: 'Tempo limite',
        timeoutDesc: 'Por quanto tempo o host pode não dar sinal de vida antes de o watchdog agir.',
        action: 'Ação',
        actionDesc:
          'O ciclo de energia segura o botão de energia por 5 segundos e depois o pressiona de novo.',
        actionReset: 'Reset',
        actionPower: 'Ciclo de energia',
        cooldown: 'Intervalo',
        cooldownDesc: 'O tempo mínimo entre duas ações.',
        maxPerHour: 'Ações por hora',
        maxPerHourDesc: 'O máximo de ações em uma hora.',
        pingHost: 'Endereço de ping',
        pingHostDesc:
          'O endereço IP do host. Uma resposta conta como sinal de vida. Deixe vazio para não fazer ping.',
        pingHostInvalid: 'Digite um endereço IPv4 ou IPv6.',
        minutes: 'min',
        save: 'Salvar',
        saved: 'Salvo',
        state: 'Detector',
        status: {
          off: 'Desligado',
          watching: 'Observando',
          hostOff: 'Host desligado',
          captureOff: 'Captura HDMI desligada',
          cooldown: 'Em intervalo',
          capped: 'Limite por hora atingido',
          acting: 'Agindo'
        },
        signal: 'Sinal HDMI',
        yes: 'Sim',
        no: 'Não',
        led: 'LED de energia',
        on: 'Aceso',
        off: 'Apagado',
        ledNotConnected: 'Não conectado',
        ping: 'Ping',
        pingNotSet: 'Não definido',
        pingReply: 'Responde',
        pingNoReply: 'Sem resposta',
        lastChange: 'Última mudança de imagem',
        never: 'Nunca',
        actsIn: 'Age em',
        actionsLastHour: 'Ações na última hora',
        duration: '{{minutes}} min {{seconds}} s',
        log: 'Registro',
        noLog: 'O watchdog ainda não agiu.',
        refresh: 'Atualizar',
        reasonFrozen: 'A imagem não mudou',
        reasonNoSignal: 'Sem sinal HDMI',
        stuckFor: 'sem sinal de vida por {{duration}}',
        pressFailed: 'O pressionamento falhou: {{error}}',
        noScreenshot: 'Sem captura de tela',
        failed: 'A operação do watchdog falhou',
        powerNeedsLed:
          'O ciclo de energia precisa de "LED de energia conectado" no menu de energia.',
        noLedConfirmTitle: 'Ativar o watchdog sem o LED de energia?',
        noLedConfirmDesc:
          'A placa não consegue ver quando o host está desligado, então o trata como sempre ligado. Se você desligar o host, o watchdog pressiona reset quando o tempo limite passar. Conecte o LED de energia para evitar isso.',
        noLedConfirmOk: 'Ativar',
        cancel: 'Cancelar'
      },
      netboot: {
        title: 'Boot pela rede',
        description:
          'Inicializar o host pela rede: iPXE e um menu das imagens no KVM pelo link de rede USB, ou netboot.xyz por proxy DHCP na LAN.',
        addon: 'dnsmasq e arquivos de boot',
        addonDesc:
          'Instalados em /data: dnsmasq do Alpine, iPXE e netboot.xyz das suas versões publicadas, cada um conferido pelo seu checksum.',
        install: 'Instalar',
        installing: 'Instalando. Isso pode levar alguns minutos.',
        uninstall: 'Desinstalar',
        uninstallConfirm: 'Desligar o boot pela rede e remover o dnsmasq e os arquivos de boot?',
        needsData: 'O boot pela rede precisa de uma imagem IronKVM com a partição /data montada.',
        usb: 'No link de rede USB',
        usbDesc:
          'Enquanto o link de rede USB está ligado, o dnsmasq o atende no lugar do udhcpd. O host recebe seu único endereço sem roteador e sem servidor DNS, o iPXE da sua arquitetura e um menu das imagens ISO no KVM.',
        linkOff: 'O link de rede USB está desligado. Ligue-o em Dispositivo, Rede USB.',
        menuUrl: 'Menu',
        leases: 'Concessão do host',
        noLeases: 'Nenhuma ainda',
        netbootxyzNote:
          'O netboot.xyz no menu carrega da internet, que o link USB não alcança. O host precisa de internet em outra porta de rede.',
        lan: 'Proxy DHCP na LAN',
        lanDesc:
          'Responde aos clientes PXE da LAN com o netboot.xyz, que depois carrega seu menu da internet. Nunca distribui endereços e não serve as imagens do KVM.',
        lanWarning:
          'O netboot.xyz é oferecido a todo cliente PXE desta LAN, não só ao host. Ligue isto apenas em uma rede que você controla.',
        lanConfirm: 'Ligar o proxy DHCP na LAN?',
        lanInterface: 'LAN',
        running: 'Em execução',
        stopped: 'Parado',
        images: 'Imagens no menu',
        noImages: 'Nenhuma imagem ISO no diretório de imagens.',
        boots: 'Boots recentes',
        noBoots: 'O host ainda não buscou nada.',
        log: 'Log do dnsmasq',
        refresh: 'Atualizar',
        okBtn: 'Confirmar',
        cancelBtn: 'Cancelar',
        failed: 'Falha na operação de boot pela rede'
      },
      about: {
        title: 'Sobre o NanoKVM',
        information: 'Informação',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versão do Aplicativo',
        applicationTip: 'Versão do aplicativo web NanoKVM',
        image: 'Versão da Imagem',
        imageTip: 'Versão da imagem do sistema NanoKVM',
        kernel: 'Versão do Kernel',
        kernelTip: 'Versão do kernel Linux em execução agora',
        deviceKey: 'Chave do Dispositivo',
        videoMemory: 'Memória de Vídeo',
        videoMemoryTip:
          'Memória reservada para a captura de vídeo. Ela não é compartilhada com o resto do sistema.',
        videoMemoryGenerations_one:
          '{{count}} sessão anterior do NanoKVM está retendo memória de vídeo',
        videoMemoryGenerations_other:
          '{{count}} sessões anteriores do NanoKVM estão retendo memória de vídeo',
        videoMemoryReboot: 'Reinicie para recuperá-la.',
        community: 'Comunidade',
        hostname: 'Nome do Host',
        hostnameUpdated: 'Nome do host atualizado. Reinicie para aplicar.',
        ipType: {
          Wired: 'Com Fio',
          Wireless: 'Sem Fio',
          Other: 'Outro'
        },
        hostnameInvalid:
          'Use letras, dígitos e hifens, até 63 por parte separada por pontos. Sem hífen no início ou no fim de uma parte.',
        hostnameFailed: 'Falha ao alterar o nome do host'
      },
      appearance: {
        title: 'Aparência',
        display: 'Exibição',
        language: 'Idioma',
        languageDesc: 'Selecione o idioma da interface',
        webTitle: 'Título da Web',
        webTitleDesc: 'Personalizar o título da página web',
        menuBar: {
          title: 'Barra de Menu',
          mode: 'Modo de exibição',
          modeDesc: 'Exibir barra de menu na tela',
          modeOff: 'Desligado',
          modeAuto: 'Ocultar automaticamente',
          modeAlways: 'Sempre visível',
          keyboardLedStatus: 'Indicadores de bloqueio do teclado',
          keyboardLedStatusDesc:
            'Exibir o estado de Num Lock, Caps Lock e Scroll Lock do computador remoto',
          icons: 'Ícones do submenu',
          iconsDesc: 'Exibir ícones de submenus na barra de menu'
        }
      },
      keyboardLedStatus: {
        groupLabel: 'Estado dos bloqueios do teclado remoto',
        indicatorLabel: '{{label}}: {{state}}',
        numLock: 'Num Lock',
        numLockShort: 'Num',
        capsLock: 'Caps Lock',
        capsLockShort: 'Caps',
        scrollLock: 'Scroll Lock',
        scrollLockShort: 'Scr',
        on: 'Ativado',
        off: 'Desativado',
        unknown: 'Desconhecido'
      },
      device: {
        title: 'Dispositivo',
        oled: {
          title: 'OLED',
          description: 'Desligar tela OLED após',
          brightness: 'Brilho do OLED',
          brightnessDescription: 'Um nível mais baixo prolonga a vida útil da tela',
          brightnessLevels: {
            '64': 'Mínimo',
            '96': 'Baixo',
            '128': 'Médio',
            '160': 'Alto',
            '207': 'Padrão',
            '255': 'Máximo'
          },
          0: 'Nunca',
          15: '15 seg',
          30: '30 seg',
          60: '1 min',
          180: '3 min',
          300: '5 min',
          600: '10 min',
          1800: '30 min',
          3600: '1 hora'
        },
        ssh: {
          description: 'Habilitar acesso remoto SSH',
          tip: 'Defina uma senha forte antes de habilitar (Conta - Mudar Senha)'
        },
        advanced: 'Configurações Avançadas',
        cpuFreq: {
          title: 'Frequência da CPU',
          description: 'Defina o clock da CPU aplicado na próxima inicialização',
          tip: 'A CPU inicia a 850 MHz e é especificada para 1000 MHz. Um novo valor é aplicado na próxima inicialização, não com o sistema em execução. 1000 MHz está dentro da especificação; a temperatura fica bem dentro dos limites em qualquer uma das opções.',
          running: 'Em execução: {{mhz}} MHz',
          rebootToApply: 'reinicie para aplicar',
          rebootConfirm: 'Reiniciar agora para aplicar {{mhz}} MHz?'
        },
        swap: {
          title: 'Swap',
          disable: 'Desativar',
          description: 'Defina o tamanho do arquivo de swap',
          tip: 'Habilitar esta função pode encurtar a vida útil do seu cartão SD!'
        },
        zram: {
          title: 'Swap comprimido (zram)',
          description: 'Swap na RAM comprimida, em vez de no cartão SD',
          tip: 'O zram mantém o swap fora do cartão SD, então não causa desgaste. Não há swap em disco por trás dele: se o zram encher, o kernel encerra um processo em vez de paginar lentamente. O limite de memória define quanta RAM o zram pode usar.',
          unavailable: 'Os módulos do kernel não estão instalados neste dispositivo',
          inactive: 'Habilitado, mas o dispositivo não iniciou',
          active: 'Ativo - {{used}} de {{total}}, {{ratio}}x',
          off: 'Desligado',
          detail: {
            algorithm: 'Algoritmo: {{algorithm}}',
            memory: 'Memória usada: {{used}} de {{limit}}',
            memoryNoLimit: 'Memória usada: {{used}}, sem limite definido',
            counters:
              'Páginas trocadas: entrada {{in}}, saída {{out}} (todos os dispositivos de swap, desde a inicialização)'
          }
        },
        mouseJiggler: {
          title: 'Movimentador de Mouse',
          description: 'Impedir que o host remoto entre em suspensão',
          disable: 'Desativar',
          absolute: 'Modo Absoluto',
          relative: 'Modo Relativo'
        },
        mdns: {
          description: 'Habilitar serviço de descoberta mDNS',
          tip: 'Desligue se não for necessário'
        },
        hdmi: {
          description: 'Habilitar saída HDMI/monitor',
          idleTimeoutTitle: 'Tempo limite de captura inativa',
          idleTimeoutDescription: 'Parar a captura HDMI após não haver visualizadores ativos por',
          minutes: 'min'
        },
        autostart: {
          title: 'Configurações de scripts de inicialização automática',
          description:
            'Gerencia scripts que são executados automaticamente na inicialização do sistema',
          new: 'Novo',
          deleteConfirm: 'Tem certeza de que deseja excluir este arquivo?',
          yes: 'Sim',
          no: 'Não',
          scriptName: 'Nome do script de inicialização automática',
          scriptContent: 'Conteúdo do script de inicialização automática',
          settings: 'Configurações'
        },
        hidOnly: 'Modo Somente-HID',
        hidOnlyDesc: 'Pare de emular dispositivos virtuais, mantendo apenas o controle básico HID',
        disk: 'Disco Virtual',
        diskDesc: 'Montar U-disk virtual no host remoto',
        network: 'Rede Virtual',
        networkDesc: 'Montar placa de rede virtual no host remoto',
        usbNetwork: {
          description:
            'Um link de rede privado com o host remoto pelo cabo USB. O host recebe um endereço sem gateway e sem DNS, então não consegue alcançar sua LAN pelo NanoKVM.',
          off: 'Desligado',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (para hosts sem NCM)',
          rndis: 'RNDIS (não é mais oferecido)',
          rndisNote: 'Este link usa RNDIS, que não é mais oferecido. Escolha NCM ou ECM.',
          subnet: 'Sub-rede',
          subnetDesc:
            'Uma rede IPv4 privada, de /24 a /30. O NanoKVM usa o primeiro endereço e o host o segundo.',
          addresses: 'NanoKVM: {{board}}, host: {{host}}',
          invalidSubnet: 'Digite uma sub-rede como 172.31.255.0/30.',
          apply: 'Aplicar',
          confirm: 'Reconectar o dispositivo USB?',
          reenumerate:
            'Aplicar reconstrói a conexão USB. O host perde o teclado, o mouse e o disco virtual por alguns segundos.'
        },
        audio: 'Alto-falante Virtual',
        audioDesc:
          'Apresenta uma placa de som USB ao host remoto, para que você possa ouvi-lo. O host deve selecioná-la como dispositivo de saída. Alterar isso reconstrói a conexão USB.',
        audioNote: 'O áudio está disponível nos dois modos H.264 (WebRTC e Direto), não no MJPEG',
        console: 'Console Serial',
        consoleDesc:
          'Apresenta uma porta serial USB ao host remoto, para entrar neste NanoKVM quando a rede estiver inacessível',
        consoleTip:
          'Quem controla o host remoto recebe um prompt de login deste NanoKVM. Defina uma senha forte antes de habilitar (Conta - Mudar Senha).',
        endpoints: {
          title: 'Endpoints USB',
          used: '{{used}} de {{total}} em uso',
          cost: 'usa {{cost}}',
          needs: 'precisa de {{cost}}',
          full: 'Não há endpoints USB suficientes. Desative outra coisa primeiro.',
          inactive:
            'Ativado, mas sem funcionar: o controlador USB ficou sem endpoints. Desative outro dispositivo e este inicia na hora.',
          explain:
            'O controlador USB tem um número fixo de endpoints de entrada, e esta é a contagem deles. Se houver mais dispositivos habilitados do que cabem, o teclado e o mouse são mantidos e o resto é desativado.',
          error: 'Não foi possível acessar o dispositivo. Tente novamente.',
          fitTogether: 'Cabem juntos: {{sets}}'
        },
        reboot: 'Reiniciar',
        rebootDesc: 'Tem certeza de que deseja reiniciar o NanoKVM?',
        okBtn: 'Sim',
        cancelBtn: 'Não',
        rebootFailed: 'Falha ao reiniciar'
      },
      network: {
        title: 'Rede',
        wifi: {
          title: 'Wi-Fi',
          description: 'Configurar Wi-Fi',
          apMode: 'O modo AP está ativado, conecte-se ao Wi-Fi escaneando o QR code',
          connect: 'Conectar Wi-Fi',
          connectDesc1: 'Digite o SSID da rede e a senha',
          connectDesc2: 'Digite a senha para entrar nesta rede',
          disconnect: 'Tem certeza de que deseja desconectar a rede?',
          failed: 'Falha na conexão, tente novamente.',
          ssid: 'Nome',
          password: 'Senha',
          joinBtn: 'Entrar',
          confirmBtn: 'OK',
          cancelBtn: 'Cancelar'
        },
        tls: {
          description: 'Habilitar protocolo HTTPS',
          tip: 'Atenção: O uso de HTTPS pode aumentar a latência, especialmente com o modo de vídeo MJPEG.',
          restarting: 'Reiniciando o servidor do dispositivo, isso leva cerca de dois minutos...',
          waiting: 'Aguardando o dispositivo responder novamente...',
          waitingHttp: 'Voltando para http. Recarregue esta página se ela não abrir sozinha.',
          failed: 'Não foi possível alterar a configuração HTTPS',
          enableConfirm: 'Ativar HTTPS?',
          disableConfirm: 'Desativar HTTPS?',
          confirmDesc:
            'Isso encerra sua sessão e reinicia o servidor do dispositivo, o que leva cerca de dois minutos. Depois a página abre {{url}}.',
          confirmOk: 'Continuar',
          confirmCancel: 'Cancelar'
        },
        ethernet: {
          title: 'Endereço IP',
          description: 'Configure como o NanoKVM obtém seu endereço na rede cabeada',
          dhcp: 'DHCP',
          manual: 'Manual',
          networkDetails: 'Detalhes da rede',
          interface: 'Interface',
          ipAddress: 'Endereço IP',
          subnetMask: 'Máscara de sub-rede',
          router: 'Roteador',
          save: 'Aplicar',
          invalidAddress: 'Informe um endereço IP válido',
          invalidMask: 'Informe uma máscara de sub-rede válida, como 255.255.255.0 ou 24',
          invalidRouter: 'Informe um endereço de roteador válido',
          addressRequired: 'Um endereço IP é obrigatório',
          maskRequired: 'Uma máscara de sub-rede é obrigatória',
          applyTitle: 'Alterar o endereço do NanoKVM?',
          applyWarning:
            'A conexão com esta página será perdida. O NanoKVM aplica o novo endereço e espera {{seconds}} segundos para você alcançá-lo nesse endereço. Alcançá-lo mantém a alteração. Se nada o alcançar, o NanoKVM restaura as configurações anteriores.',
          applyConfirm: 'Aplicar',
          applyCancel: 'Cancelar',
          applyFailed: 'Falha ao aplicar o endereço',
          trialTitle: 'Aguardando confirmação',
          trialDhcp: 'O NanoKVM está solicitando um endereço por DHCP.',
          trialStatic: 'O NanoKVM agora está em {{address}}.',
          trialInstruction:
            'Abra o NanoKVM no novo endereço e entre, se ele pedir. Alcançá-lo ali mantém a alteração. Se nada alcançar o NanoKVM em {{seconds}} segundos, ele restaura as configurações anteriores.',
          trialOpen: 'Abrir o novo endereço',
          trialKeep: 'Manter estas configurações',
          trialKept: 'O novo endereço está salvo',
          trialKeepFailed: 'Falha ao manter as configurações',
          trialGone: 'A alteração já foi desfeita. Tente novamente.',
          unsaved: 'Alterações não salvas'
        },
        dns: {
          title: 'DNS',
          description: 'Configurar servidores DNS para o NanoKVM',
          mode: 'Modo',
          dhcp: 'DHCP',
          manual: 'Manual',
          add: 'Adicionar DNS',
          save: 'Salvar',
          invalid: 'Digite um endereço IP válido',
          noDhcp: 'Nenhum DNS DHCP está disponível no momento',
          saved: 'Configurações de DNS salvas',
          saveFailed: 'Falha ao salvar as configurações de DNS',
          unsaved: 'Alterações não salvas',
          maxServers: 'Máximo de {{count}} servidores DNS permitido',
          dnsServers: 'Servidores DNS',
          dhcpServersDescription: 'Os servidores DNS são obtidos automaticamente via DHCP',
          manualServersDescription: 'Os servidores DNS podem ser editados manualmente',
          networkDetails: 'Detalhes da rede',
          interface: 'Interface',
          ipAddress: 'Endereço IP',
          subnetMask: 'Máscara de sub-rede',
          router: 'Roteador',
          none: 'Nenhum'
        }
      },
      vpn: {
        loading: 'Carregando...',
        okBtn: 'Sim',
        cancelBtn: 'Não',
        restart: 'Reiniciar o {{name}}?',
        stop: 'Parar o {{name}}?',
        stopDesc:
          'O daemon para agora. Iniciar na inicialização é uma opção separada e continua como está.',
        update: 'Atualizar o {{name}} para {{version}}?',
        updateDesc: 'O daemon reinicia se estiver em execução. O login é mantido.',
        notInstall: 'O {{name}} não está instalado.',
        install: 'Instalar',
        installing: 'Instalando',
        installFailed: 'Falha na instalação',
        retry: 'Tentar novamente',
        notRunning: 'O {{name}} não está em execução. Inicie-o para continuar.',
        run: 'Iniciar',
        boot: 'Iniciar na inicialização',
        bootDesc: 'Inicia o {{name}} quando o KVM é ligado.',
        enable: 'Habilitar {{name}}',
        control: 'Servidor de controle',
        connected: 'Conectado',
        disconnected: 'Não conectado',
        deviceName: 'Nome do dispositivo',
        deviceIP: 'IP do dispositivo',
        account: 'Conta',
        version: 'Versão',
        uptime: 'Tempo ativo',
        peers: 'Peers',
        noPeers: 'Nenhum peer ainda.',
        online: 'Online',
        offline: 'Offline',
        memory: 'Memória',
        daemonRss: 'Daemon',
        group: 'Grupo de complementos',
        high: 'limitado acima de {{size}}',
        max: 'encerrado pelo kernel acima de {{size}}',
        noGroup: 'Não há grupo de memória de complementos nesta placa.',
        uninstall: 'Desinstalar {{name}}',
        uninstallDesc:
          'Tem certeza de que deseja desinstalar o {{name}}? O login permanece na placa.',
        blocked:
          'O {{other}} está em execução ou inicia na inicialização. Só uma VPN funciona por vez: pare o {{other}} e desative a inicialização automática dele primeiro.',
        swap: {
          title: 'Memória swap',
          tip: 'Se faltar memória ao daemon, tente habilitar a memória swap. Isso define o tamanho do arquivo de swap como 256MB por padrão, o que pode ser ajustado em "Configurações > Dispositivo".'
        }
      },
      tailscale: {
        title: 'Tailscale',
        retry: 'Por favor, atualize e tente novamente. Ou tente instalar manualmente',
        download: 'Baixar o',
        package: 'pacote de instalação',
        unzip: 'e descompacte-o',
        upTailscale: 'Fazer upload do tailscale para o diretório NanoKVM /usr/bin/',
        upTailscaled: 'Fazer upload do tailscaled para o diretório NanoKVM /usr/sbin/',
        refresh: 'Atualizar página atual',
        notLogin:
          'O dispositivo ainda não foi vinculado. Por favor, faça login e vincule este dispositivo à sua conta.',
        urlPeriod: 'Esta URL é válida por 10 minutos',
        login: 'Login',
        loginSuccess: 'Login Bem-sucedido',
        logout: 'Sair',
        logoutDesc: 'Tem certeza de que deseja sair?'
      },
      netbird: {
        title: 'NetBird',
        notLogin:
          'Este dispositivo ainda não entrou em uma rede NetBird. Entre com uma chave de configuração ou faça login com SSO.',
        setupKey: 'Chave de configuração',
        setupKeyPlaceholder: 'Cole uma chave de configuração do painel do NetBird',
        join: 'Entrar',
        or: 'ou',
        sso: 'Login com SSO',
        urlPeriod: 'Esta URL é válida por 10 minutos',
        loginSuccess: 'Login Bem-sucedido',
        logout: 'Cancelar registro',
        logoutDesc:
          'Cancelar o registro remove este peer da sua conta NetBird e apaga a configuração dele aqui. Para entrar de novo é preciso uma chave de configuração ou um login SSO, e o peer pode receber um novo IP. Continuar?'
      },
      update: {
        title: 'Verificar Atualizações',
        queryFailed: 'Falha ao obter a versão',
        updateFailed: 'Falha na atualização. Por favor, tente novamente.',
        isLatest: 'Você já tem a versão mais recente.',
        available: 'Uma atualização está disponível. Tem certeza de que deseja atualizar agora?',
        updating: 'Atualização iniciada. Por favor, aguarde...',
        confirm: 'Confirmar',
        cancel: 'Cancelar',
        preview: 'Prévia das Atualizações',
        previewDesc: 'Tenha acesso antecipado a novos recursos e melhorias',
        previewTip:
          'Esteja ciente de que as versões de prévia podem conter bugs ou funcionalidade incompleta!',
        customServer: {
          title: 'Servidor de atualização personalizado',
          desc: 'Verifique e baixe atualizações online de um servidor especificado',
          invalidUrl:
            'Insira um diretório de servidor HTTP ou HTTPS válido, sem parâmetros de consulta, fragmentos ou latest.json.',
          loadFailed: 'Não foi possível carregar a configuração do servidor de atualização.',
          saveFailed: 'Não foi possível salvar a configuração do servidor de atualização.',
          saved: 'Configuração do servidor de atualização salva.',
          save: 'Salvar',
          confirmTitle: 'Usar um servidor de atualização personalizado?',
          confirmDesc:
            'O SHA-512 apenas verifica se o pacote corresponde ao manifesto fornecido por este servidor. Ele não comprova que o pacote seja uma versão oficial do NanoKVM. Um servidor com falha ou mal-intencionado pode inutilizar o dispositivo, causar perda de dados ou comprometer o sistema.',
          confirm: 'Usar mesmo assim',
          useSipeed: 'Usar o servidor oficial da Sipeed',
          previewDisabled:
            'As atualizações de prévia ficam indisponíveis enquanto um servidor de atualização personalizado estiver ativado.'
        },
        offline: {
          title: 'Atualizações off-line',
          desc: 'Atualização através do pacote de instalação local',
          upload: 'Upload',
          checksumPlaceholder: 'Soma de verificação SHA-256 (opcional)',
          invalidChecksum: 'A soma de verificação SHA-256 deve conter 64 caracteres hexadecimais.',
          checksumMismatch: 'A verificação SHA-256 falhou. O pacote pode estar corrompido.',
          invalidName: 'Formato de nome de arquivo inválido. Faça download das versões do GitHub.',
          updateFailed: 'Falha na atualização. Por favor, tente novamente.'
        },
        updateTo: 'Atualizar para {{version}}',
        updateConfirmDesc:
          'O dispositivo instala a atualização e reinicia seu servidor. Esta página recarrega quando o servidor voltar.'
      },
      account: {
        title: 'Conta',
        webAccount: 'Nome da Conta Web',
        role: 'Função',
        roles: { admin: 'Administrador', user: 'Usuário' },
        password: 'Senha',
        updateBtn: 'Alterar',
        logoutBtn: 'Sair',
        logoutDesc: 'Tem certeza de que deseja sair?',
        okBtn: 'Sim',
        cancelBtn: 'Não',
        users: {
          title: 'Usuários',
          create: 'Criar Usuário',
          enabled: 'Habilitado',
          disabled: 'Desabilitado',
          deviceOwner: 'Dono do dispositivo',
          resetPassword: 'Redefinir Senha',
          delete: 'Excluir',
          deleteConfirm: 'Excluir este usuário e revogar todas as sessões dele?',
          created: 'Usuário criado',
          deleted: 'Usuário excluído',
          passwordUpdated: 'Senha atualizada',
          loadFailed: 'Falha ao carregar os usuários',
          saveFailed: 'Falha ao salvar o usuário',
          deleteFailed: 'Falha ao excluir o usuário'
        }
      },
      apiKeys: {
        title: 'Chaves de API',
        description:
          'Uma chave age como o seu dono, com a função desse usuário. Envie-a como Authorization: Bearer <key> para métricas e a API, ou como X-Auth-Token para o Redfish.',
        name: 'Nome',
        namePlaceholder: 'Para que serve a chave, como prometheus',
        nameRequired: 'Dê um nome à chave',
        nameTooLong: 'O nome pode ter no máximo 64 caracteres',
        unnamed: '(sem nome)',
        create: 'Criar Chave',
        created: 'Criada',
        owner: 'Dono',
        empty: 'Nenhuma chave de API',
        newKeyTitle: 'Sua nova chave de API',
        newKeyWarning:
          'Copie a chave agora. Ela não é armazenada e não pode ser exibida novamente. Se você perdê-la, revogue-a e crie outra.',
        copy: 'Copiar',
        copied: 'Copiado',
        copyFailed: 'Falha ao copiar. Copie manualmente.',
        done: 'Concluído',
        revoke: 'Revogar',
        revokeConfirmTitle: 'Revogar esta chave de API?',
        revokeConfirmDesc: 'Tudo o que usa "{{name}}" para de funcionar imediatamente.',
        revoked: 'Chave de API revogada',
        loadFailed: 'Falha ao carregar as chaves de API',
        createFailed: 'Falha ao criar a chave de API',
        revokeFailed: 'Falha ao revogar a chave de API',
        cancelBtn: 'Cancelar'
      }
    },
    picoclaw: {
      title: 'PicoClaw Assistente',
      empty: 'Abra o painel e inicie uma tarefa para começar.',
      inputPlaceholder: 'Descreva o que você deseja que PicoClaw faça',
      newConversation: 'Nova conversa',
      processing: 'Processando...',
      agent: {
        defaultTitle: 'Assistente geral',
        defaultDescription: 'Ajuda geral sobre bate-papo, pesquisa e espaço de trabalho.',
        kvmTitle: 'Controle remoto',
        kvmDescription: 'Opera o host remoto por meio de NanoKVM.',
        switched: 'Função de agente trocada',
        switchFailed: 'Falha ao mudar de função de agente'
      },
      send: 'Enviar',
      cancel: 'Cancelar',
      status: {
        connecting: 'Conectando ao gateway...',
        connected: 'Sessão PicoClaw conectada',
        disconnected: 'Sessão PicoClaw desconectada',
        stopped: 'Solicitação de parada enviada',
        runtimeStarted: 'Runtime do PicoClaw iniciado',
        runtimeStartFailed: 'Falha ao iniciar o runtime do PicoClaw',
        runtimeStopped: 'Runtime do PicoClaw interrompido',
        runtimeStopFailed: 'Falha ao parar o runtime do PicoClaw',
        controlSwitchedToMCP: 'Controle transferido para o serviço MCP externo'
      },
      connection: {
        runtime: {
          checking: 'Verificando',
          restoring: 'Restoring PicoClaw',
          ready: 'Runtime pronto',
          stopped: 'Runtime interrompido',
          blockedByMCP: 'O controle MCP externo está ativo',
          readyBlockedByMCP:
            'The runtime is running, but external MCP currently controls device input.',
          readyWithoutControl:
            'The runtime is running. Grant PicoClaw device control before reconnecting.',
          unavailable: 'Runtime indisponível',
          configError: 'Erro de configuração'
        },
        transport: {
          connecting: 'Conectando',
          connected: 'Conectado',
          disconnected: 'Disconnected',
          reconnect: 'Reconnect',
          reconnectDescription: 'Reconnect to the running PicoClaw session.',
          reconnectBlocked: 'PicoClaw needs device control before reconnecting.'
        },
        run: {
          idle: 'Inativo',
          busy: 'Ocupado'
        }
      },
      message: {
        toolAction: 'Ação',
        observation: 'Observação',
        screenshot: 'Captura de tela'
      },
      overlay: {
        locked: 'PicoClaw está controlando o dispositivo. A entrada manual está pausada.'
      },
      control: {
        picoclaw: 'Controle do dispositivo: PicoClaw',
        picoclawDescription: 'PicoClaw can write keyboard and mouse input. Manual input may pause.',
        mcp: 'Controle do dispositivo: MCP externo',
        mcpDescription: 'External MCP can write to the device. PicoClaw will not take over input.',
        off: 'Controle do dispositivo: desativado',
        offDescription:
          'AI will not write keyboard or mouse input. Manual control remains available.',
        transitioning: 'Device control: switching',
        transitioningDescription: 'Device control is syncing. Please wait.',
        grant: 'Conceder controle',
        release: 'Liberar',
        releasing: 'Releasing...',
        switching: 'Switching...',
        releasingLabel: 'Device control: releasing',
        releasingDescription:
          'Device control is being returned. PicoClaw has stopped current writes.',
        granted: 'Controle do PicoClaw concedido',
        released: 'Controle do PicoClaw liberado',
        grantFailed: 'Falha ao conceder controle ao PicoClaw',
        releaseFailed: 'Falha ao liberar controle do PicoClaw',
        grantConfirmTitle: 'Alternar controle do dispositivo para PicoClaw?',
        grantConfirmDesc: 'As gravações de dispositivo do MCP externo serão interrompidas.'
      },
      install: {
        install: 'Instalar PicoClaw',
        installing: 'Instalando PicoClaw',
        success: 'PicoClaw instalado com sucesso',
        failed: 'Falha ao instalar PicoClaw',
        uninstalling: 'Desinstalando o runtime...',
        uninstalled: 'Runtime desinstalado com sucesso.',
        uninstallFailed: 'Falha na desinstalação.',
        requiredTitle: 'PicoClaw não está instalado',
        requiredDescription: 'Instale o PicoClaw antes de iniciar o runtime do PicoClaw.',
        progressDescription: 'PicoClaw está sendo baixado e instalado.',
        stages: {
          preparing: 'Preparando',
          downloading: 'Baixando',
          extracting: 'Extraindo',
          verifying: 'Verificando',
          installing: 'Instalando',
          installed: 'Instalado',
          install_timeout: 'Tempo limite esgotado',
          install_failed: 'Falhou'
        }
      },
      model: {
        requiredTitle: 'A configuração do modelo é necessária',
        requiredDescription: 'Configure o modelo PicoClaw antes de usar o chat PicoClaw.',
        docsTitle: 'Guia de configuração',
        docsDesc: 'Modelos e protocolos suportados',
        menuLabel: 'Configurar modelo',
        modelIdentifier: 'Identificador do modelo',
        modelIdentifierPlaceholder: 'openai/gpt-5.4',
        apiBase: 'API Base URL',
        apiBasePlaceholder: 'https://api.example.com/v1',
        apiKey: 'Chave API',
        apiKeyPlaceholder: 'Insira a chave API do modelo',
        save: 'Salvar',
        saving: 'Salvando',
        saved: 'Configuração do modelo salva',
        saveFailed: 'Falha ao salvar a configuração do modelo',
        invalid: 'Identificador do modelo, API Base URL e chave API são obrigatórios'
      },
      uninstall: {
        menuLabel: 'Desinstalar',
        confirmTitle: 'Desinstalar PicoClaw',
        confirmContent:
          'Tem certeza de que deseja desinstalar PicoClaw? Isso excluirá o executável e todos os arquivos de configuração.',
        confirmOk: 'Desinstalar',
        confirmCancel: 'Cancelar'
      },
      history: {
        title: 'Histórico',
        loading: 'Carregando sessões...',
        emptyTitle: 'Ainda sem histórico',
        emptyDescription: 'As sessões anteriores de PicoClaw aparecerão aqui.',
        loadFailed: 'Falha ao carregar o histórico da sessão',
        deleteFailed: 'Falha ao excluir sessão',
        deleteConfirmTitle: 'Excluir sessão',
        deleteConfirmContent: 'Tem certeza de que deseja excluir "{{title}}"?',
        deleteConfirmOk: 'Excluir',
        deleteConfirmCancel: 'Cancelar',
        messageCount_one: '{{count}} mensagem',
        messageCount_other: '{{count}} mensagens',
        messageCount: '{{count}} mensagens'
      },
      config: {
        startRuntime: 'Iniciar PicoClaw',
        stopRuntime: 'Parar PicoClaw'
      },
      start: {
        enableConfirmTitle: 'Transferir o controle para o PicoClaw?',
        enableConfirmDesc: 'Iniciar o PicoClaw desabilitará o serviço MCP externo.',
        enableConfirmOk: 'Iniciar PicoClaw',
        enableConfirmCancel: 'Cancelar',
        title: 'Iniciar PicoClaw',
        description: 'Inicie o runtime para começar a usar o assistente PicoClaw.',
        switchFromMCP: 'Switch to PicoClaw and start',
        takeoverAndStart: 'Take over and start'
      }
    },
    error: {
      title: 'Encontramos um problema',
      refresh: 'Atualizar',
      panel: 'Esta parte da página parou de funcionar',
      retry: 'Tentar novamente'
    },
    fullscreen: {
      toggle: 'Alternar Tela Cheia'
    },
    input: {
      disconnected: 'Teclado e mouse não estão conectados',
      disconnectedTls:
        'O navegador recusou a conexão segura que transporta o teclado e o mouse, e faz isso sem perguntar. O certificado gerado por este dispositivo ainda não é confiável. Abra este endereço em uma nova aba, aceite o certificado e recarregue. Instalar o certificado é a solução definitiva.',
      disconnectedNever:
        'Não foi possível abrir a conexão que transporta o teclado e o mouse. O resto da página funciona porque não a utiliza. Verifique se nada entre você e o dispositivo a está bloqueando.',
      disconnectedDropped:
        'A conexão que transporta o teclado e o mouse caiu e não voltou. Ela se reconecta sozinha após uma reinicialização; se isso persistir, recarregue a página.',
      hidDisabled: 'O HID está desativado neste dispositivo (/boot/disable_hid).',
      keyFailed: 'Não foi possível enviar a tecla.'
    },
    speaker: { title: 'Alto-falante', unmute: 'Ativar som', mute: 'Silenciar' },
    menu: {
      collapse: 'Recolher Menu',
      expand: 'Expandir Menu'
    },
    ion: {
      checking: 'Verificando a memória de vídeo antes de iniciar a transmissão...',
      warn: 'A memória de vídeo está baixa. Uma única reinicialização do servidor a esgotaria. Reinicie quando for conveniente.',
      criticalTitle: 'Memória de vídeo insuficiente para iniciar a transmissão',
      criticalBody:
        'Iniciar o vídeo esgotaria a memória reservada e pararia o servidor. Todas as outras funções continuam funcionando, incluindo o controle de energia e a reinicialização. Só uma reinicialização do NanoKVM recupera essa memória.',
      criticalContinue: 'Iniciar o vídeo mesmo assim',
      criticalReboot: 'Reiniciar o NanoKVM',
      criticalRebooting: 'Reiniciando...'
    }
  }
};

export default pt_br;
