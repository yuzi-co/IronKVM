const pt_br = {
  translation: {
    feedback: {
      enabled: '{{name}} ativado',
      disabled: '{{name}} desativado',
      failed: 'A solicitação falhou. Tente novamente.',
      network: 'Não foi possível alcançar o dispositivo. Verifique a conexão e tente novamente.',
      saved: 'Salvo',
      timeout: 'O dispositivo demorou demais para responder. Tente novamente.'
    },
    common: {
      copy: 'Copiar',
      copied: 'Copiado',
      copyFailed: 'Não foi possível copiar. Selecione o texto e copie manualmente.',
      notUpdating: 'Sem atualização: a última consulta falhou.',
      off: 'Desligado',
      running: 'Em execução',
      save: 'Salvar',
      cancel: 'Cancelar',
      delete: 'Excluir',
      remove: 'Remover'
    },
    head: {
      desktop: 'Área de Trabalho Remota',
      login: 'Login',
      changePassword: 'Mudar Senha',
      terminal: 'Terminal',
      wifi: 'Wi-Fi'
    },
    auth: {
      share: {
        opening: 'Abrindo a tela compartilhada...',
        failed: 'Este link é inválido, expirou ou foi revogado.',
        toLogin: 'Entrar com uma conta'
      },
      passwordChanged: 'Senha alterada. Entre com a nova senha.',
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
          'Para redefinir as senhas, pressione e segure o botão BOOT no IronKVM por 10 segundos.',
        reset3: 'Conta padrão da Web:',
        reset4: 'Conta padrão SSH:',
        change1: 'Por favor, note que esta ação irá alterar as seguintes senhas:',
        change2: 'Senha de login da Web',
        change3: 'Senha root do sistema (senha de login SSH)',
        change4: 'Para redefinir as senhas, pressione e segure o botão BOOT no IronKVM.',
        resetDocs: 'Para ver o passo a passo, consulte a documentação do hardware:',
        hardwareDocs: 'Wiki do Sipeed NanoKVM'
      }
    },
    wifi: {
      title: 'Wi-Fi',
      description: 'Configurar Wi-Fi para o IronKVM',
      success: 'Por favor, verifique o status da rede do IronKVM e visite o novo endereço IP.',
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
        verifyBtn: 'Verificar',
        tooManyAttempts: 'Muitas senhas incorretas. Aguarde um momento e tente novamente.'
      },
      ssidRequired: 'Digite o nome da rede, até 32 caracteres',
      passwordLength: 'A senha tem de 8 a 63 caracteres. Deixe vazia para uma rede aberta.',
      passwordOptional: 'Senha (vazia para uma rede aberta)',
      lost: 'A placa parou de responder. Ela pode ter entrado na rede e fechado seu hotspot de configuração. Se o hotspot voltar, a conexão falhou: conecte-se a ele de novo e tente outra vez.',
      done: 'Configuração concluída. Reconecte este dispositivo à sua rede habitual e abra a placa em seu novo endereço.'
    },
    screen: {
      viewOnly: 'Somente visualizar',
      viewOnlyTip:
        'Esta aba deixa de enviar teclado e mouse ao host. Scripts, o jiggler do mouse e outros espectadores não são afetados.',
      viewOnlyOff: 'Desativar somente visualizar',
      viewOnlyBlocked: 'Somente visualizar está ativo, nada foi enviado ao host',
      pauseHidden: 'Pausar com a aba oculta',
      pauseHiddenTip:
        'Para o vídeo e o som alguns segundos depois que esta aba é ocultada e os retoma quando você volta.',
      screenshot: 'Captura de tela',
      screenshotTip: 'Salva a tela do host como PNG no tamanho total de captura.',
      screenshotFailed: 'Falha na captura de tela',
      streamInfo: {
        toggle: 'Info do stream',
        toggleTip:
          'Mostra abaixo da imagem uma linha com a conexão, a resolução, a taxa de quadros, a taxa de bits e a latência do stream, como este navegador as mede.',
        title: 'Status do stream',
        connecting: 'Conectando',
        connected: 'Conectado',
        reconnecting: 'Reconectando',
        failed: 'Falhou',
        paused: 'Pausado',
        noSignal: 'Sem sinal',
        source: 'Fonte',
        sourceTip: 'O modo da entrada HDMI que o dispositivo captura',
        stream: 'Stream',
        streamTip: 'O tamanho da imagem como chega a este navegador',
        received: 'recebidos',
        decoded: 'decodificados',
        dropped: 'descartados',
        board: 'dispositivo',
        boardTip: 'Quadros por segundo que o dispositivo captura e codifica',
        bitrateTip: 'Dados de vídeo que este navegador recebe',
        latency: 'Latência',
        latencyCapture:
          'Da captura no dispositivo até a exibição aqui, pela hora de captura que o servidor envia com cada quadro.',
        latencyEncoder:
          'Da saída do codificador até o desenho aqui, medida pelo relógio do dispositivo. Captura e codificação não estão incluídas.',
        latencyEstimate:
          'Estimativa: metade da ida e volta da rede mais o buffer e a decodificação deste navegador. Captura e codificação no dispositivo não estão incluídas, então a latência real é maior.',
        latencyEstimateSum:
          'Soma de partes medidas: {{board}} ms no dispositivo da captura ao envio, {{network}} ms no caminho até aqui (metade da ida e volta da rede), {{receive}} ms de buffer e decodificação neste navegador e {{display}} ms de espera pela tela (meia atualização).',
        latencyClockOffset:
          'Corrigido pela diferença entre o relógio do dispositivo e o deste computador, medida em {{offset}} ms (dispositivo menos este computador).',
        latencyClockAssumed:
          'Supõe-se que o relógio do dispositivo coincide com o deste computador.',
        latencyArrival:
          'Os quadros chegam aqui {{ms}} ms após o codificador; o resto é decodificação e desenho neste navegador.',
        rttTip: 'Ida e volta da rede até o dispositivo',
        loss: 'Perda',
        lossTip: 'Pacotes perdidos no último segundo',
        transportTip: 'Como o vídeo chega a este navegador'
      },
      stream: {
        ok: 'imagem OK',
        noSignal: 'sem sinal',
        failed: 'falha no stream'
      },
      codecNoWebrtcHevc: 'Este navegador não recebe H.265 via WebRTC',
      codecNoHevc: 'Este navegador não decodifica H.265',
      codecNote:
        'A placa tem um só codificador, então isso muda o stream para todos os espectadores.',
      codec: 'Codec',
      updateFailed: 'A configuração não foi aplicada',
      scale: 'Escala',
      title: 'Tela',
      video: 'Modo de Vídeo',
      videoDirectTips: 'Ative HTTPS em "Configurações > Dispositivo" para usar este modo',
      resolution: 'Resolução',
      aspect: 'Proporção de tela',
      aspectKeep: 'Automático (manter a proporção da origem)',
      aspectStretch: 'Esticar para a resolução',
      aspectTips:
        'Automático mantém o formato da tela do host na altura selecionada: uma tela de 1920x1200 é enviada como 1728x1080. Esticar preenche a resolução selecionada e distorce telas que não são 16:9.',
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
      qualityLossless: 'Máxima',
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
      },
      directConnectionFailed: 'Falha na conexão do fluxo de vídeo'
    },
    keyboard: {
      close: 'Fechar',
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
      autoTap: {
        title: 'Toque automático',
        tip: 'Pressiona uma tecla repetidamente para abrir um menu do firmware ou de inicialização. O dispositivo continua pressionando quando esta aba é fechada.',
        key: 'Tecla',
        otherKey: 'Outra tecla',
        interval: 'Intervalo (ms)',
        timeout: 'Parar após (s)',
        firstPress: 'Primeiro pressionamento',
        firstPressNone: 'Nada',
        firstPressPower: 'Botão liga/desliga',
        firstPressReset: 'Botão de reset',
        stopOnInput: 'Parar na minha próxima tecla ou clique',
        stopOnVideo: 'Parar quando o modo de vídeo mudar',
        stopOnVideoTip:
          'Um menu do firmware costuma ter uma resolução diferente da do sistema. Uma reinicialização pode mudar o modo mais de uma vez, e alguns firmwares mantêm o mesmo modo.',
        start: 'Iniciar',
        stop: 'Parar'
      },
      macro: {
        title: 'Macros',
        empty: 'Nenhuma macro salva',
        record: 'Gravar uma macro',
        recordTip:
          'Grava suas teclas, o mouse, as pausas entre eles e os pressionamentos de liga/desliga ou reset, até você clicar em Concluir.',
        play: 'Reproduzir',
        edit: 'Editar',
        delete: 'Excluir',
        deleteConfirm: 'Excluir a macro "{{name}}"?',
        recording: 'Gravando uma macro. Passos: {{count}}',
        recordingFull: 'A macro atingiu o limite de passos.',
        finish: 'Concluir',
        discard: 'Descartar',
        editorTitle: 'Macro',
        name: 'Nome',
        namePlaceholder: 'Nome da macro',
        summary: 'Passos: {{count}}. Duração: cerca de {{seconds}} s',
        addPause: 'Adicionar uma pausa',
        addPower: 'Adicionar um pressionamento de liga/desliga',
        addReset: 'Adicionar um pressionamento de reset',
        removeStep: 'Remover o passo',
        playNow: 'Reproduzir agora',
        save: 'Salvar',
        saved: 'Macro salva',
        adminOnly: 'Somente um administrador pode salvar macros.',
        noSteps: 'A macro não tem passos.',
        noName: 'Dê um nome à macro.',
        stepKeys: 'Pressionar {{keys}}',
        stepRelease: 'Soltar as teclas',
        stepPointerAt: 'Ponteiro em {{x}}%, {{y}}%',
        stepPointerBy: 'Mover o ponteiro {{x}}, {{y}}',
        stepButtons: 'botões: {{buttons}}',
        stepNoButtons: 'nenhum botão',
        stepScroll: 'rolagem {{amount}}',
        stepPause: 'Pausa (ms)',
        stepPower: 'Botão liga/desliga, {{ms}} ms',
        stepReset: 'Botão de reset, {{ms}} ms',
        buttonLeft: 'esquerdo',
        buttonRight: 'direito',
        buttonMiddle: 'do meio'
      },
      sequence: {
        autoTapRunning: 'Toque automático {{name}}. Pressionamentos: {{taps}}',
        macroRunning: 'Macro {{name}}. Passo {{step}} de {{total}}',
        elapsed: '{{elapsed}} s de {{timeout}} s',
        stop: 'Parar',
        alreadyRunning: 'Um toque automático ou uma macro já está em execução. Pare-o primeiro.',
        endedAutoTap: 'Toque automático {{name}}: {{reason}}',
        endedMacro: 'Macro {{name}}: {{reason}}',
        reasons: {
          finished: 'execução concluída',
          stopped: 'execução interrompida',
          timeout: 'execução interrompida no limite de tempo',
          input: 'execução interrompida pela sua tecla ou clique',
          video: 'execução interrompida, o modo de vídeo mudou',
          shutdown: 'execução interrompida, o servidor reiniciou',
          busy: 'falha, outro controlador está usando o teclado',
          hid_error: 'falha, o host não aceitou a entrada',
          atx_error: 'falha, o pressionamento de liga/desliga ou reset não funcionou'
        }
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
        sendFailed: 'Não enviado: a conexão de entrada caiu',
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
        saveFailed: 'Falha ao salvar a tecla líder',
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
      jiggler: 'Movimentador de mouse',
      keyJiggler: 'Pressionador de tecla',
      keyJigglerF15: 'Tecla F15',
      keyJigglerShift: 'Tecla Shift',
      keyJigglerCtrl: 'Tecla Ctrl',
      keyJigglerF15Tip: 'F15 é a menos intrusiva: nenhum sistema ou app comum a usa',
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
      scrollUp: 'Igual a este computador',
      scrollDown: 'Invertido (rolagem natural)',
      speed: 'Velocidade da roda de rolagem',
      fast: 'Rápido',
      slow: 'Lento',
      requestPointer:
        'Usando modo relativo. Por favor, clique na área de trabalho para obter o ponteiro do mouse.',
      resetHid: 'Redefinir HID',
      hidOnly: {
        switchFailed: 'Não foi possível trocar o modo. Verifique a conexão e tente novamente.',
        title: 'Modo somente HID',
        desc: 'Se o seu mouse e teclado pararem de responder e a redefinição de HID não ajudar, pode ser um problema de compatibilidade entre o IronKVM e o dispositivo. Tente habilitar o modo Somente-HID para melhor compatibilidade.',
        tip1: 'Habilitar o modo Somente-HID irá desmontar o U-disk virtual e a rede virtual',
        tip2: 'No modo Somente-HID, a montagem de imagem está desativada',
        rebuild: 'Trocar de modo reconstrói a conexão USB. O IronKVM não reinicia',
        enable: 'Habilitar modo Somente-HID',
        disable: 'Desabilitar modo Somente-HID'
      },
      resetHidDone: 'HID USB redefinido',
      resetHidFailed: 'Falha ao redefinir o HID USB'
    },
    image: {
      driveLoaded: 'imagem carregada',
      driveWarning: 'veja os avisos',
      warning: {
        missing:
          'O arquivo de imagem foi apagado. O host continua lendo a cópia antiga até você ejetá-la.',
        writable: 'Leitura e gravação: o host pode alterar esta imagem.',
        tooBigForCd: 'Grande demais para a unidade de CD ({{size}}, limite {{max}}). Use o disco.',
        tooSmallForCd: 'Pequena demais para a unidade de CD ({{size}}). Use o disco.',
        empty: 'O arquivo está vazio, provavelmente por um upload ou download que falhou.'
      },
      delete: 'Excluir',
      inUse: 'Em uso. Ejete antes de excluir.',
      retry: 'Tentar novamente',
      loadFailed: 'Não foi possível carregar a lista de imagens',
      readOnlyLocked: 'Ejete o disco para alterar isso. Vale quando uma imagem é inserida.',
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
        usb1: 'Conecte o IronKVM ao seu computador via USB.',
        usb2: 'Certifique-se de que o disco virtual está montado (Configurações - Disco Virtual).',
        usb3: 'Abra o disco virtual no seu computador e copie o arquivo de imagem para o diretório raiz do disco virtual.',
        scp1: 'Certifique-se de que o IronKVM e seu computador estão na mesma rede local.',
        scp2: 'Abra um terminal no seu computador e use o comando SCP para fazer upload do arquivo de imagem para o diretório /data no IronKVM.',
        scp3: 'Exemplo: scp seu-caminho-da-imagem root@seu-ip-nanokvm:/data',
        tfCard: 'Cartão TF',
        tf1: 'Este método é suportado em sistemas Linux',
        tf2: 'Remova o cartão TF do IronKVM (para a versão FULL, desmonte a caixa primeiro).',
        tf3: 'Insira o cartão TF em um leitor de cartão e conecte-o ao seu computador.',
        tf4: 'Copie o arquivo de imagem para o diretório /data no cartão TF.',
        tf5: 'Insira o cartão TF no IronKVM.'
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
      close: 'Fechar',
      empty: 'Nenhum script ainda. Envie um arquivo .sh ou .py para executá-lo na placa.',
      loadFailed: 'Falha ao carregar os scripts',
      uploaded: 'Script enviado',
      uploadFailed: 'Falha ao enviar o script',
      started: 'Script iniciado em segundo plano',
      deleteFailed: 'Falha ao excluir o script',
      waitLimit: 'Aguardando o script terminar, por até {{minutes}} minutos.',
      timedOut:
        'O script levou mais de {{minutes}} minutos e esta página parou de esperar. Ele pode ainda estar em execução na placa.'
    },
    terminal: {
      invalidBaud: 'Esta taxa de transmissão não é suportada.',
      invalidPort: 'Informe um caminho de dispositivo em /dev, como /dev/ttyS1.',
      invalidSettings: 'Configurações da porta serial inválidas. Este é o shell da própria placa.',
      disconnected: 'Desconectado. Pressione Enter para reconectar.',
      title: 'Terminal',
      nanokvm: 'Terminal IronKVM',
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
      no: 'Não',
      yes: 'Sim',
      deleteConfirm: 'Excluir este endereço salvo?',
      delete: 'Excluir',
      wake: 'Acordar',
      rename: 'Renomear',
      showMac: 'Mostrar endereço MAC',
      showName: 'Mostrar nome',
      requestFailed: 'Não foi possível alcançar o dispositivo para enviar o comando',
      deleteFailed: 'Falha ao excluir',
      renameFailed: 'Falha ao renomear',
      title: 'Wake-on-LAN',
      sending: 'Enviando comando...',
      sent: 'Comando enviado',
      input: 'Por favor, digite o MAC',
      ok: 'Ok',
      add: 'Adicionar dispositivo',
      edit: 'Editar dispositivo',
      name: 'Nome',
      mac: 'Endereço MAC',
      scan: 'Procurar na rede',
      scanFailed: 'Não foi possível varrer a rede',
      scanEmpty: 'Nenhum host encontrado',
      scanPick: 'Escolha um host',
      ip: 'Endereço IP do host (para verificar se está ligado, opcional)',
      broadcast: 'Endereço de broadcast ou sub-rede (opcional)',
      port: 'Pacote de ativação',
      portFrame: 'Quadro Ethernet (padrão)',
      secureOn: 'Senha SecureOn (opcional)',
      secureOnKept: 'Salva. Deixe em branco para mantê-la',
      secureOnRemove: 'Remover a senha',
      secureOnKeep: 'Manter a senha',
      invalidMac: 'Informe um endereço MAC válido',
      invalidIp: 'Informe um endereço IPv4 válido',
      invalidBroadcast: 'Informe um endereço IPv4 ou uma sub-rede, como 192.168.1.0/24',
      invalidSecureOn: 'A senha tem seis bytes em hexadecimal, como AA:BB:CC:DD:EE:FF',
      duplicate: 'Já existe um dispositivo com este endereço MAC ou nome',
      saveFailed: 'Falha ao salvar o dispositivo',
      save: 'Salvar',
      cancel: 'Cancelar',
      up: 'Ligado',
      down: 'Desligado',
      checking: 'Verificando...'
    },
    download: {
      uploadFailed: 'Falha no envio',
      uploadSuccess: 'Envio concluído',
      uploading: 'Enviando: {{file}}',
      downloadingPercent: 'Baixando ({{percent}}): {{file}}',
      downloading: 'Baixando: {{file}}',
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
      bootMenuPresent: '{{file}} já está no dispositivo, com o checksum correto',
      bootMenuDesc: 'Baixar a ISO do netboot.xyz, com checksum conferido, para o CD virtual'
    },
    alerts: {
      title: 'Requer atenção',
      temperature: {
        warning: 'A placa está a {{celsius}} °C. Verifique se o ar chega até ela.',
        critical: 'A placa está a {{celsius}} °C, quente demais. Dê ventilação ou desligue-a.'
      },
      storage: {
        warning:
          'Apenas {{available}} livres de {{total}} em {{path}}. Imagens grandes podem não caber.',
        critical:
          'Apenas {{available}} livres em {{path}}. Uploads, downloads e instalações de complementos vão falhar. Apague imagens de que não precisa.'
      },
      vpn: '{{name}} deve iniciar no boot, mas não está em execução, então o acesso remoto por ele está fora.',
      openVpn: 'Abrir configurações de VPN',
      stream:
        'O stream de vídeo falhou. Tente outro modo de vídeo no menu Tela ou recarregue a página.'
    },
    power: {
      resetDesc: 'Reinicia o host na hora. O trabalho não salvo é perdido.',
      powerShortDesc: 'Liga o host, ou pede ao sistema operacional que desligue (ACPI).',
      powerLongDesc: 'Força o desligamento do host sem encerrar o sistema.',
      hddLed: 'LED do disco',
      hddActive: 'Ativo',
      hddIdle: 'Inativo',
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
      nav: {
        system: 'Sistema',
        network: 'Rede',
        access: 'Acesso',
        integrations: 'Integrações',
        boot: 'Inicialização e mídia',
        browser: 'Este navegador',
        search: 'Buscar uma configuração',
        noMatch: 'Nenhuma configuração encontrada',
        locked:
          'Uma operação está em andamento. Outras páginas e o fechamento ficam indisponíveis até ela terminar.',
        vpnProvider: 'Provedor de VPN'
      },
      audit: {
        title: 'Registro de auditoria',
        log: 'Eventos',
        description:
          'Quem entrou, assumiu o controle, pressionou o botão de energia, alterou mídias ou configurações e o que os clientes de IA fizeram. Quando o registro enche, os eventos mais antigos são descartados.',
        refresh: 'Atualizar',
        export: 'Exportar',
        time: 'Hora',
        who: 'Quem',
        address: 'Endereço',
        category: 'Categoria',
        action: 'Ação',
        target: 'Alvo',
        outcome: 'Resultado',
        tier: 'Nível: {{tier}}',
        empty: 'Nenhum evento',
        searchPlaceholder: 'Buscar ação, alvo ou detalhe',
        loadFailed: 'Não foi possível carregar o registro de auditoria.',
        exportFailed: 'Não foi possível exportar o registro de auditoria.',
        saveFailed: 'Não foi possível salvar a configuração.',
        syslog: 'Enviar eventos ao syslog',
        syslogDescription:
          'Envia também cada evento ao registro do sistema, que o encaminha ao coletor syslog remoto, se houver um definido em Rede.',
        storage: 'Guardado em {{directory}}, no máximo {{size}} KB.',
        categories: {
          auth: 'Login',
          control: 'Controle',
          power: 'Energia',
          media: 'Mídia',
          settings: 'Configurações',
          macro: 'Macros',
          system: 'Sistema',
          agent: 'Agentes de IA',
        },
        outcomes: {
          ok: 'Sucesso',
          fail: 'Falha',
          denied: 'Negado'
        }
      },
      mcp: {
        tiers: {
          title: 'Níveis de capacidade',
          description:
            'Escolha o que os clientes MCP e o agente integrado podem fazer. Cada chamada é registrada no registro de auditoria com o seu nível.',
          mcp: 'Clientes MCP',
          agent: 'Agente integrado',
          observe: 'Observar',
          observeDesc: 'Capturas de tela e status',
          input: 'Entrada',
          inputDesc: 'Teclado e mouse',
          power: 'Energia',
          powerDesc: 'Pressionar os botões de energia e reset ATX',
          media: 'Mídia',
          mediaDesc: 'Inserir e ejetar mídias virtuais',
          admin: 'Administração',
          adminDesc:
            'Alterar as configurações do dispositivo (nenhuma ferramenta usa este nível ainda)',
          agentNote:
            'O agente integrado ainda não tem ferramentas de energia ou mídia, então esses dois níveis valem apenas para clientes MCP.',
          loadFailed: 'Não foi possível carregar os níveis de capacidade.',
          saveFailed: 'Não foi possível salvar os níveis de capacidade.'
        },
        keyNote:
          'O MCP usa sua própria chave de API, mostrada abaixo. Chaves da página Chaves de API não funcionam aqui.',
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
        cancelBtn: 'Cancelar',
        showKey: 'Mostrar chave',
        hideKey: 'Ocultar chave',
        regenerateKey: 'Gerar nova chave'
      },
      redfish: {
        example: 'Exemplo',
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
        copyBeforeSave: 'Copie a senha agora. Depois de salva, ela não pode ser exibida de novo.',
        noLogin:
          'O IPMI está ligado, mas nenhuma conta ativa tem senha IPMI, então ninguém consegue entrar. Defina uma abaixo.',
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
      ssh: {
        service: 'Servidor SSH',
        serviceDesc: 'Iniciar o sshd agora e a cada inicialização',
        failed: 'Não foi possível carregar as configurações de SSH',
        rootDefault: 'root ainda tem a senha de fábrica',
        rootEmpty: 'root não tem senha',
        rootWarning:
          'Qualquer pessoa que chegue ao console ou ao SSH pode entrar como root. Defina uma senha em {{account}} > {{password}}: para o dono do dispositivo ela também define a senha do root.',
        connection: 'Conexão',
        command: 'Entrar como root',
        port: 'Porta',
        viaVpn: 'Via {{name}}',
        notRunning: 'O sshd não está em execução. Ative o servidor SSH para se conectar.',
        hostKeys: 'Impressões digitais das chaves de host',
        hostKeysDesc: 'Compare-as com o que o ssh mostra na primeira conexão.',
        noHostKeys: 'Ainda não há chaves de host. O sshd as cria na primeira inicialização.',
        keys: 'Chaves autorizadas',
        keysDesc:
          'Chaves públicas que podem entrar como root. Ficam na partição de dados, então as atualizações as mantêm.',
        noKeys: 'Nenhuma chave autorizada ainda.',
        noComment: 'sem comentário',
        addPlaceholder: 'Cole uma chave pública, como o conteúdo de ~/.ssh/id_ed25519.pub',
        add: 'Adicionar chave',
        added: 'Chave adicionada',
        removed: 'Chave removida',
        deleteConfirm: 'Remover esta chave?',
        deleteConfirmDesc: 'Ela não poderá mais entrar. Sessões abertas continuam abertas.',
        invalidKey: 'Isto não é uma chave pública. Cole uma única linha de um arquivo .pub.',
        keyOptions: 'Chaves com opções como command= ou from= não são aceitas aqui.',
        duplicateKey: 'Esta chave já está autorizada.',
        lastKey:
          'A última chave não pode ser removida enquanto o acesso somente por chave estiver ativo.',
        keysOnly: 'Somente chaves',
        keysOnlyDesc:
          'Desativar o login por senha e keyboard-interactive. Sessões abertas continuam abertas.',
        keysOnlyNeedsKey:
          'Adicione primeiro uma chave autorizada, senão ninguém conseguiria entrar.',
        keysOnlyOn: 'Login por senha desativado',
        keysOnlyOff: 'Login por senha ativado',
        notHonoured:
          'O sshd desta imagem não lê esta configuração, então o login por senha continua ativo.',
        reloadFailed:
          'Salvo, mas não foi possível recarregar o sshd. Vale na próxima vez que o sshd iniciar.',
        notApplied:
          'O sshd ainda aceita senhas. Desligue e ligue o servidor SSH para aplicar a configuração.',
        changePort: 'Alterar',
        portConfirm: 'Alterar a porta SSH para {{port}}?',
        portConfirmDesc:
          'Suas sessões SSH atuais continuam abertas. Novas conexões precisam usar a porta {{port}}. Verifique se o firewall permite.',
        portChanged: 'Porta SSH alterada para {{port}}',
        portInvalid: 'Informe uma porta de 1 a 65535.',
        portReserved: 'O próprio IronKVM usa esta porta. Escolha outra.',
        portInUse: 'Outro programa no IronKVM já escuta nesta porta.',
        portNotHonoured:
          'O sshd desta imagem não lê esta configuração, então a porta continua a mesma.'
      },
      vnc: {
        address: 'Endereço',
        certHint:
          'O VeNCrypt X509Plain usa o certificado autoassinado do dispositivo, então o cliente avisa na primeira conexão. Aceite-o, ou salve o certificado pelo endereço HTTPS desta página e passe-o ao TigerVNC com -X509CA=<arquivo>.',
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
      notifications: {
        title: 'Notificações',
        service: 'Enviar notificações',
        serviceDesc:
          'Envia um alerta, com uma imagem da tela, quando algo acontece com o host ou com este KVM. Desativado por padrão.',
        snapshot: 'Anexar um instantâneo',
        snapshotDesc:
          'Um JPEG da tela acompanha o alerta onde o canal puder levá-lo. Ele não é armazenado.',
        events: 'Eventos',
        eventsDesc:
          'Escolha o que é enviado e o tempo mínimo entre dois alertas do mesmo tipo. Os alertas retidos por esse tempo são contados e informados junto com o próximo.',
        event: {
          signalLost: 'Sinal HDMI perdido',
          signalRestored: 'Sinal HDMI restaurado',
          watchdogAction: 'Watchdog pressionou reset ou energia',
          loginFailures: 'Falhas de login repetidas',
          atxPress: 'Botão de energia ou reset pressionado',
          failsafeEntered: 'Modo failsafe: o servidor iniciou sem vídeo',
          hostLogMatch: 'Frase no log do host, como um kernel panic'
        },
        cooldown: 'Tempo mínimo entre dois alertas',
        seconds: 'segundos',
        loginWithin: 'logins com falha em',
        channels: 'Canais',
        webhook: 'Webhook',
        webhookDesc: 'Envia o evento como JSON para um endereço.',
        webhookUrl: 'Endereço',
        webhookSnapshot: 'Adicionar o instantâneo ao JSON em base64',
        ntfy: 'ntfy',
        ntfyDesc: 'Publica em um tópico no ntfy.sh ou no seu próprio servidor.',
        ntfyServer: 'Servidor',
        ntfyTopic: 'Tópico',
        ntfyToken: 'Token de acesso (opcional)',
        telegram: 'Telegram',
        telegramDesc: 'Envia por um bot que você criou com o BotFather.',
        botToken: 'Token do bot',
        chatId: 'ID do chat',
        savedSecret: 'Salvo. Digite para substituir.',
        savedHint: 'Salvo para {{hint}}. Digite para substituir.',
        willRemove: 'Será removido ao salvar',
        remove: 'Remover',
        sendTest: 'Enviar teste',
        sendTestAll: 'Enviar teste para todos os canais',
        saveFirst: 'Salve suas alterações primeiro',
        testOk: 'Teste enviado em {{channel}}',
        testChannelFailed: '{{channel}}: {{error}}',
        testFailed: 'Não foi possível enviar o teste',
        save: 'Salvar',
        saved: 'Salvo',
        failed: 'Falha nas configurações de notificação',
        recent: 'Alertas recentes',
        refresh: 'Atualizar',
        noRecent: 'Nada foi enviado ainda.',
        heldBack: 'Retidos',
        lastOk: 'enviado',
        lastFailed: 'falhou'
      },
      hostLog: {
        title: 'Log do host',
        intro:
          'Receba os logs do host gerenciado pela rede: syslog e netconsole do kernel. Ambos ficam desativados até você ativá-los. As linhas ficam na memória e, se você pedir, no disco.',
        syslog: 'Syslog (UDP)',
        syslogDesc: 'RFC 3164 e RFC 5424. Aponte o rsyslog ou o syslog-ng do host para este KVM.',
        netconsole: 'Netconsole (UDP)',
        netconsoleDesc:
          'Mensagens do kernel, mesmo enquanto os discos do host estão falhando. Formatos simples e estendido.',
        port: 'Porta',
        state: {
          off: 'Desativado',
          listening: 'Escutando',
          waiting: 'Aguardando',
          error: 'Erro'
        },
        stateListening: 'Em {{addr}}',
        stateWaiting: 'Aguardando o link de rede USB',
        bind: 'Escutar em',
        bindDesc: 'A opção USB ouve apenas o host ligado por cabo a este KVM.',
        bindAll: 'Todas as interfaces',
        bindUsb: 'Somente o link de rede USB',
        sources: 'Remetentes aceitos',
        sourcesDesc:
          'Endereços ou redes, um por linha. Se vazio, só o host no link USB (ou o endereço de ping do watchdog) é ouvido quando conhecido; caso contrário, qualquer remetente.',
        inForce: 'Em vigor: {{list}}',
        inForceAny: 'Em vigor: qualquer remetente',
        memory: 'Buffer de memória',
        memoryDesc:
          'As linhas mais novas que couberem são mantidas. As mais antigas são descartadas.',
        disk: 'Manter também o log no disco',
        diskDesc:
          'Sobrevive a uma reinicialização. Grava no cartão SD, por isso fica desativado por padrão.',
        diskLimit: 'Limite de disco',
        patterns: 'Frases de alerta',
        patternsDesc:
          'Expressões regulares, uma por linha, sem distinção de maiúsculas. Uma linha correspondente envia uma notificação. Se vazio, são usadas as padrão: kernel panic, oops, BUG e out of memory.',
        save: 'Salvar',
        saved: 'Salvo',
        failed: 'Falha na requisição do log do host',
        stats:
          '{{received}} linhas recebidas, {{filtered}} ignoradas de outros remetentes, {{dropped}} descartadas por chegarem rápido demais. {{count}} linhas mantidas.',
        live: 'Log ao vivo',
        paused: 'Pausado',
        connected: 'Ao vivo',
        reconnecting: 'Reconectando',
        search: 'Buscar palavras',
        severityAll: 'Todos os níveis',
        severityError: 'Erros e mais graves',
        severityWarning: 'Avisos e mais graves',
        severityInfo: 'Info e mais graves',
        pause: 'Pausar',
        resume: 'Retomar',
        download: 'Baixar',
        clear: 'Limpar',
        cancel: 'Cancelar',
        empty: 'Nenhuma linha ainda.',
        clearConfirmTitle: 'Limpar o log do host?',
        clearConfirmDesc: 'As linhas mantidas na memória e no disco serão apagadas.'
      },
      nut: {
        title: 'Cliente NUT',
        service: 'Monitoramento do UPS',
        serviceDesc: 'Lê um UPS de um servidor NUT na rede. Com bateria baixa ou um desligamento forçado, pressiona o botão de energia do host para desligá-lo e pressiona de novo quando a rede elétrica volta.',
        notice: 'O KVM precisa continuar ligado durante a queda de energia, e o conector ATX precisa estar ligado. O protocolo NUT não é criptografado: o usuário e a senha trafegam na rede em texto puro.',
        needsLed: 'O cliente NUT precisa de "LED de energia conectado" ativado, para saber se o host está ligado.',
        ledHint: '"LED de energia conectado" está desativado, então o cliente não pressiona nenhum botão.',
        host: 'Servidor NUT',
        hostDesc: 'A máquina que executa o upsd: um NAS, roteador, placa de rede do UPS ou o próprio host.',
        hostInvalid: 'Digite um nome de host ou um endereço IP.',
        port: 'Porta',
        upsName: 'Nome do UPS',
        username: 'Usuário',
        credentialsDesc: 'Opcional. Com usuário e senha, o KVM entra como secondary, e o primary espera o host desligar antes de cortar a saída do UPS.',
        password: 'Senha',
        passwordSet: 'Há uma senha salva.',
        passwordKeep: 'Sem alteração',
        passwordClear: 'Remover senha',
        onBattery: 'Desligar após',
        onBatteryDesc: 'Minutos na bateria antes de pedir ao host que desligue. 0 desativa isso; bateria baixa e desligamento forçado sempre desligam o host.',
        forceOff: 'Forçar desligamento após',
        forceOffDesc: 'Minutos após o pressionamento de desligar antes de segurar o botão de energia por 5 segundos. 0 desativa isso.',
        powerOn: 'Ligar quando a rede elétrica voltar',
        powerOnDesc: 'Pressiona o botão de energia quando o UPS volta à rede elétrica e o host está desligado após a queda. Um host desligado de propósito é deixado em paz.',
        minCharge: 'Carga mínima',
        minChargeDesc: 'Carga da bateria que o UPS precisa atingir antes. 0 significa qualquer uma. O UPS precisa informar battery.charge.',
        delay: 'Atraso',
        delayDesc: 'Por quanto tempo a rede elétrica precisa estar de volta antes do pressionamento.',
        minutes: 'min',
        secondsUnit: 's',
        save: 'Salvar',
        saved: 'Salvo',
        failed: 'Não foi possível carregar ou salvar as configurações do NUT',
        state: 'Estado',
        status: {
          off: 'Desligado',
          unconfigured: 'Não configurado',
          connecting: 'Conectando',
          ok: 'Conectado',
          unreachable: 'Inacessível'
        },
        upsStateLabel: 'UPS',
        upsState: {
          online: 'Na rede elétrica',
          onBattery: 'Na bateria',
          lowBattery: 'Battery baixa',
          shutdown: 'Desligamento',
          unknown: 'Desconhecido'
        },
        charge: 'Carga',
        runtime: 'Autonomia',
        onBatteryFor: 'Na bateria há',
        hostPower: 'Energia do host',
        hostState: {
          on: 'Ligado',
          off: 'Desligado',
          unknown: 'Desconhecido'
        },
        login: 'Login',
        loginSecondary: 'Secondary: o primary espera',
        loginReadOnly: 'Somente leitura',
        loginRefused: 'Recusado ({{code}})',
        activity: 'Atividade',
        phase: {
          idle: 'Ocioso',
          shuttingDown: 'Desligando o host',
          waitingForMains: 'Aguardando a rede elétrica para ligar',
          needsLed: 'Precisa do LED de energia'
        },
        powerOnIn: 'Ligar em',
        lastPoll: 'Última consulta',
        never: 'Nunca',
        duration: '{{minutes}} min {{seconds}} s',
        events: 'Eventos',
        noEvents: 'Nada ainda.',
        event: {
          unreachable: 'Servidor NUT inacessível',
          reachable: 'Servidor NUT acessível de novo',
          onBattery: 'UPS na bateria',
          mainsBack: 'A rede elétrica voltou',
          shutdown: 'Desligamento do host pressionado',
          forceOff: 'Host desligado à força',
          hostOff: 'O host está desligado',
          powerOn: 'Ligar o host pressionado',
          pressFailed: 'Falha ao pressionar o botão'
        },
        reason: {
          fsd: 'desligamento forçado',
          lowBattery: 'bateria baixa',
          onBattery: 'tempo na bateria'
        },
        barTip: 'UPS lido do servidor NUT: estado, carga da bateria e autonomia',
        runtimeShort: '{{minutes}} min'
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
        actionReset: 'Redefinir',
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
      media: {
        title: 'Mídia virtual',
        description:
          'Configuração da janela Mídia da barra de ferramentas. Montar imagens, adicioná-las e escolher o conjunto do Ventoy é feito na janela.',
        ejectFirst:
          'O disco do Ventoy está em uma unidade. Ejete-o na janela Mídia para desinstalar.'
      },
      netboot: {
        title: 'Boot pela rede',
        isoDownload: 'Baixar',
        updateFiles: 'Arquivos de inicialização (inicialização pela rede)',
        updateIso: 'ISO do menu de inicialização (CD virtual)',
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
        title: 'Sobre o IronKVM',
        information: 'Informação',
        ip: 'IP',
        mdns: 'mDNS',
        application: 'Versão do Aplicativo',
        applicationTip: 'Versão do aplicativo web IronKVM',
        image: 'Versão da Imagem',
        imageTip:
          'Imagem de cartão do IronKVM e a imagem de sistema do NanoKVM na qual ela se baseia',
        kernel: 'Versão do Kernel',
        kernelTip: 'Versão do kernel Linux em execução agora',
        deviceKey: 'Chave do Dispositivo',
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
        hostnameFailed: 'Falha ao alterar o nome do host',
        editHostname: 'Editar nome do host',
        docs: 'Documentação',
        hardware: 'Hardware',
        hardwareFaq: 'Perguntas frequentes do hardware',
        disclaimer:
          'IronKVM: firmware comunitário reforçado para o Sipeed NanoKVM. Sem vínculo com a Sipeed.',
        basedOn: 'baseado no NanoKVM {{version}}',
        licenses: 'Licenças',
        licensesDescription: 'Licenças do software desta imagem',
        licensesUnavailable: 'Esta imagem não inclui um manifesto de licenças.',
        licenseTexts: 'Textos de licença ({{count}})'
      },
      preferences: {
        title: 'Preferências'
      },
      performance: {
        title: 'Desempenho',
        memory: {
          title: 'Memória',
          description: 'RAM, swap e o que os usa. Atualiza a cada poucos segundos.',
          ram: 'RAM',
          of: '{{used}} de {{total}}',
          available: '{{available}} disponíveis',
          availableLow:
            'Apenas {{available}} disponíveis. Serviços podem ficar lentos ou ser parados.',
          swap: 'Swap',
          swapFile: 'Arquivo de swap',
          zram: 'Swap comprimido',
          zramRam: '{{ram}} na RAM',
          consumers: 'Principais consumidores',
          addons: 'Complementos',
          addonsTip:
            'O grupo de memória em que o Tailscale e o NetBird rodam, em relação ao seu limite.',
          video: 'Memória de Vídeo',
          videoTip:
            'Memória reservada para a captura de vídeo. Ela não é compartilhada com o resto do sistema.',
          videoGenerations_one:
            '{{count}} sessão anterior do IronKVM está retendo memória de vídeo',
          videoGenerations_other:
            '{{count}} sessões anteriores do IronKVM estão retendo memória de vídeo',
          videoReboot: 'Reinicie para recuperá-la.',
          videoUnmeasured: '{{size}} reservados, uso não mensurável',
          largestFree: 'Maior bloco livre {{size}}',
          peak: 'pico {{size}}',
          cma: 'Memória contígua (CMA)',
          cmaTip:
            'Memória que o kernel reserva para buffers que precisam ser contíguos, como os quadros de vídeo entre o redimensionador e os codificadores. Enquanto não está alocada, o resto do sistema pode usá-la.'
        }
      },
      appearance: {
        thisBrowser: 'Este navegador',
        thisBrowserDesc:
          'Salvo apenas neste navegador. Outros navegadores mantêm as próprias configurações.',
        deviceWide: 'Dispositivo',
        deviceWideDesc: 'Salvo no dispositivo. Vale para todos que o abrem.',
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
        sections: {
          video: 'Vídeo',
          usb: 'USB',
          frontPanel: 'Painel frontal',
          roomMic: 'Microfone do ambiente'
        },
        roomMic: {
          allow: 'Permitir o microfone do ambiente',
          description:
            'Os espectadores podem ouvir o ambiente pelo microfone embutido do KVM. Todos que assistem veem quando está ligado.',
          unavailable: 'Indisponível neste kernel',
          gain: 'Ganho do microfone'
        },
        cpuFreq: {
          title: 'Frequência da CPU',
          description: 'Defina o clock da CPU aplicado na próxima inicialização',
          tip: 'A CPU inicia a 850 MHz e é especificada para 1000 MHz. Um novo valor é aplicado na próxima inicialização, não com o sistema em execução. 1000 MHz está dentro da especificação; a temperatura fica bem dentro dos limites em qualquer uma das opções.',
          running: 'Em execução: {{mhz}} MHz',
          rebootToApply: 'reinicie para aplicar',
          rebootConfirm: 'Reiniciar agora para aplicar {{mhz}} MHz?'
        },
        videoClock: {
          title: 'Clocks de vídeo',
          description: 'Perfil de clock dos codificadores e do escalonador de vídeo',
          tip: 'Rápido roda o núcleo do codec, o escalonador e o barramento do codec a 594, 396 e 450 MHz; Padrão nos valores de fábrica do chip, 396, 300 e 360 MHz. Rápido reduz a latência do stream em cerca de 4 ms (H.265) e 3 ms (H.264), sem aumento de temperatura mensurável. A mudança vale quando o vídeo começar de novo.',
          standard: 'Padrão',
          fast: 'Rápido',
          rates: 'Codec {{codec}} · escalonador {{vpss}} · barramento {{bus}} MHz',
          pending: 'vale quando o vídeo começar de novo'
        },
        swap: {
          title: 'Swap',
          disable: 'Desativar',
          description: 'Defina o tamanho do arquivo de swap',
          tip: 'Habilitar esta função pode encurtar a vida útil do seu cartão SD!',
          active: 'Ativo - {{used}} de {{total}}',
          inactive: 'Configurado, mas sem uso'
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
        hidOnly: 'Modo Somente-HID',
        hidOnlyDesc: 'Pare de emular dispositivos virtuais, mantendo apenas o controle básico HID',
        disk: 'Disco Virtual',
        diskDesc: 'Montar U-disk virtual no host remoto',
        network: 'Rede Virtual',
        networkDesc: 'Montar placa de rede virtual no host remoto',
        usbNetwork: {
          boardAddress: 'IronKVM:',
          hostAddress: 'Host:',
          description:
            'Um link de rede privado com o host remoto pelo cabo USB. O host recebe um endereço sem gateway e sem DNS, então não consegue alcançar sua LAN pelo IronKVM.',
          mode: 'Protocolo',
          ncm: 'NCM (Linux, macOS, Windows 11)',
          ecm: 'ECM (para hosts sem NCM)',
          rndis: 'RNDIS (não é mais oferecido)',
          rndisNote: 'Este link usa RNDIS, que não é mais oferecido. Escolha NCM ou ECM.',
          subnet: 'Sub-rede',
          subnetDesc:
            'Uma rede IPv4 privada, de /24 a /30. O IronKVM usa o primeiro endereço e o host o segundo.',
          invalidSubnet: 'Digite uma sub-rede como 172.31.255.0/30.',
          apply: 'Aplicar',
          confirm: 'Reconectar o dispositivo USB?',
          reenumerate:
            'Aplicar reconstrói a conexão USB. O host perde o teclado, o mouse e o disco virtual por alguns segundos.'
        },
        usbIdentity: {
          title: 'Identidade USB',
          description: 'Os IDs de fabricante e produto e os nomes que o host remoto vê para este dispositivo.',
          live: 'O host vê:',
          bootOverride: '/boot/usb.vid ou /boot/usb.pid está definido. Vale enquanto o padrão estiver selecionado.',
          vendorId: 'ID do fabricante',
          productId: 'ID do produto',
          manufacturer: 'Fabricante',
          product: 'Nome do produto',
          serial: 'Número de série',
          serialDesc: 'De 12 a 64 letras e dígitos. Deixe vazio para manter o da placa.',
          serialPlaceholder: 'O da placa',
          invalid: 'Verifique os campos destacados.',
          apply: 'Aplicar',
          restore: 'Restaurar padrão',
          confirm: 'Alterar a identidade USB?',
          restoreConfirm: 'Restaurar a identidade USB padrão?',
          warning: 'O host vê um novo dispositivo USB e o enumera de novo: teclado, mouse e disco virtual caem por alguns segundos, e o host pode instalar drivers.',
          error: 'Não foi possível ler a identidade USB.',
          presets: {
            default: 'Padrão',
            logitech: 'Teclado Logitech',
            microsoft: 'Teclado Microsoft',
            dell: 'Teclado Dell',
            custom: 'Personalizado'
          }
        },
        audio: 'Alto-falante Virtual',
        audioDesc:
          'Apresenta uma placa de som USB ao host remoto, para que você possa ouvi-lo. O host deve selecioná-la como dispositivo de saída. Alterar isso reconstrói a conexão USB.',
        audioNote: 'O áudio está disponível nos dois modos H.264 (WebRTC e Direto), não no MJPEG',
        console: 'Console Serial',
        consoleDesc:
          'Apresenta uma porta serial USB ao host remoto, para entrar neste IronKVM quando a rede estiver inacessível',
        consoleTip:
          'Quem controla o host remoto recebe um prompt de login deste IronKVM. Defina uma senha forte antes de habilitar (Conta - Mudar Senha).',
        usbApply: {
          changed: 'Alterado',
          discard: 'Descartar',
          pending: 'As alterações ainda não foram aplicadas.'
        },
        endpoints: {
          title: 'Slots USB',
          free: '{{free}} de {{total}} livres',
          slots: 'Slots: {{count}}',
          full: 'Slots USB livres insuficientes. Desligue outra coisa primeiro.',
          inactive:
            'Ligado, mas sem funcionar: o controlador USB ficou sem slots. Desligue outro dispositivo e este inicia na hora.',
          explain:
            'O controlador USB tem um número fixo de slots (endpoints de entrada), e o teclado e o mouse sempre ocupam alguns. Se houver mais dispositivos ligados do que cabem, o teclado e o mouse são mantidos e o resto é desligado.',
          error: 'Não foi possível acessar o dispositivo. Tente novamente.',
          fitTogether: 'Cabem juntos: {{sets}}'
        },
        reboot: 'Reiniciar',
        rebootDesc: 'Tem certeza de que deseja reiniciar o IronKVM?',
        okBtn: 'Sim',
        cancelBtn: 'Não',
        rebootFailed: 'Falha ao reiniciar'
      },
      network: {
        title: 'Rede',
        wifi: {
          disconnectBtn: 'Desconectar',
          disconnectWarning:
            'Se você acessa o IronKVM por esta rede Wi-Fi, esta página perderá a conexão.',
          disconnected: 'Wi-Fi desconectado',
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
          description: 'Configure como o IronKVM obtém seu endereço na rede cabeada',
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
          applyTitle: 'Alterar o endereço do IronKVM?',
          applyWarning:
            'A conexão com esta página será perdida. O IronKVM aplica o novo endereço e espera {{seconds}} segundos para você alcançá-lo nesse endereço. Alcançá-lo mantém a alteração. Se nada o alcançar, o IronKVM restaura as configurações anteriores.',
          applyConfirm: 'Aplicar',
          applyCancel: 'Cancelar',
          applyFailed: 'Falha ao aplicar o endereço',
          trialTitle: 'Aguardando confirmação',
          trialDhcp: 'O IronKVM está solicitando um endereço por DHCP.',
          trialStatic: 'O IronKVM agora está em {{address}}.',
          trialInstruction:
            'Abra o IronKVM no novo endereço e entre, se ele pedir. Alcançá-lo ali mantém a alteração. Se nada alcançar o IronKVM em {{seconds}} segundos, ele restaura as configurações anteriores.',
          trialOpen: 'Abrir o novo endereço',
          trialKeep: 'Manter estas configurações',
          trialKept: 'O novo endereço está salvo',
          trialKeepFailed: 'Falha ao manter as configurações',
          trialGone: 'A alteração já foi desfeita. Tente novamente.',
          unsaved: 'Alterações não salvas'
        },
        ipv6: {
          title: 'IPv6',
          description: 'Define como o IronKVM obtém o endereço IPv6 na rede cabeada',
          slaac: 'SLAAC',
          dhcpv6: 'DHCPv6',
          static: 'Manual',
          off: 'Desligado',
          address: 'Endereço IPv6',
          prefix: 'Tamanho do prefixo',
          router: 'Roteador',
          apply: 'Aplicar',
          saved: 'Configurações de IPv6 aplicadas e salvas',
          loadFailed: 'Falha ao ler as configurações de IPv6',
          saveFailed: 'Falha ao aplicar as configurações de IPv6',
          unsupported: 'Este kernel não tem IPv6',
          noDhcpv6: 'DHCPv6 não está disponível: esta imagem não tem cliente DHCPv6',
          invalid:
            'Informe um endereço IPv6 válido, um tamanho de prefixo de 16 a 126 e o endereço do roteador (ou deixe vazio)',
          offNote:
            'O IPv6 está desligado na interface cabeada. A interface web continua acessível por IPv4.',
          keepsIpv4:
            'O endereço IPv4 não muda. Se a aplicação falhar, as configurações de IPv6 anteriores são restauradas.',
          useIpv4:
            'Esta página está aberta por IPv6. Abra o IronKVM pelo endereço IPv4 para alterar as configurações de IPv6.',
          current: 'Endereços IPv6 atuais',
          disabled: 'Desligado',
          noAddress: 'Ainda sem endereço IPv6',
          scopeLink: 'link-local',
          scopeUla: 'privado',
          scopeGlobal: 'global'
        },
        webrtc: {
          title: 'WebRTC por redirecionamento de porta',
          description:
            'Permite que espectadores fora da sua rede acessem o vídeo por um redirecionamento de porta no roteador.',
          warning:
            'Redirecionar uma porta expõe este KVM à internet. Uma VPN como Tailscale ou NetBird continua sendo o acesso recomendado. Redirecione apenas a porta UDP abaixo, nunca a porta da interface web.',
          udpPort: 'Porta UDP fixa',
          udpPortHint:
            'Todos os espectadores compartilham esta porta. Deixe vazio para desativar. Redirecione-a no roteador para este dispositivo.',
          publicAddress: 'Endereço público',
          publicAddressHint:
            'O endereço IP ou o nome DNS dinâmico em que seu roteador responde. Requer a porta UDP.',
          apply: 'Aplicar',
          saved: 'Salvo. Novos espectadores usam as novas configurações.',
          loadFailed: 'Falha ao ler as configurações de WebRTC',
          saveFailed: 'Falha ao salvar as configurações de WebRTC',
          invalidPort: 'A porta deve ser um número de 1024 a 65535',
          needsPort: 'Um endereço público requer uma porta UDP',
          active: 'Os espectadores compartilham a porta UDP {{port}}',
          inactive: 'Cada espectador usa sua própria porta',
          offered: 'Oferecido aos espectadores: {{addresses}}',
          cost:
            'Em uma porta compartilhada a placa envia um pacote por vez, o que pode aumentar o uso de CPU com vários espectadores e taxas de quadros altas.'
        },
        dns: {
          title: 'DNS',
          description: 'Configurar servidores DNS para o IronKVM',
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
        },
        syslog: {
          title: 'Log remoto',
          description:
            'Envia os logs do sistema, do kernel e do IronKVM para um coletor syslog via UDP.',
          placeholder: 'host ou host:porta, UDP, porta padrão 514',
          save: 'Salvar',
          turnOff: 'Desativar',
          test: 'Enviar mensagem de teste',
          sent: 'Enviado <v>{{message}}</v>. Procure por ela no coletor.',
          forwarding: 'Encaminhando para <v>{{target}}</v>',
          local: 'Somente local',
          savedButLocal: '<v>{{target}}</v> salvo, mas o logger em execução mantém os logs locais',
          savedButForwarding:
            '<v>{{target}}</v> salvo, mas o logger em execução encaminha para <v>{{active}}</v>',
          offButForwarding:
            'Desativado, mas o logger em execução ainda encaminha para <v>{{active}}</v>',
          unsupported: 'O logger desta imagem ignora a configuração; atualize para usá-la',
          loadFailed: 'Falha ao ler a configuração de log',
          saveFailed: 'Falha ao salvar a configuração de log',
          testFailed: 'Falha ao enviar a mensagem de teste',
          metrics: 'Métricas',
          metricsDesc:
            'O Prometheus pode coletar a URL abaixo com um token bearer de <link>Chaves de API</link>.',
          metricsUrl: 'URL de métricas',
          errors: {
            empty: 'Informe um host ou host:porta',
            long: 'Longo demais para um nome de host',
            ipv6: 'Coloque o endereço IPv6 entre colchetes, como [fd00::1]:514',
            brackets: 'O endereço entre colchetes não é um endereço IPv6',
            host: 'Use apenas letras, dígitos, pontos e hifens, sem começar com ponto ou hífen',
            port: 'A porta deve ser um número de 1 a 65535'
          }
        },
        allowlist: {
          title: 'Redes permitidas',
          description:
            'Endereços que podem acessar a interface web, a API, o VNC e o IPMI. Um endereço ou CIDR por linha, como 192.168.1.0/24. Deixe vazio para permitir qualquer endereço.',
          everyone: 'Qualquer endereço pode se conectar.',
          restricted:
            'Somente as redes da lista podem se conectar. O próprio dispositivo e o link de rede USB são sempre permitidos.',
          placeholder: '192.168.1.0/24',
          yourAddress: 'Seu endereço: {{client}}',
          addMine: 'Adicionar meu endereço',
          save: 'Salvar',
          allowAll: 'Permitir qualquer endereço',
          loadFailed: 'Falha ao carregar as redes permitidas',
          saveFailed: 'Falha ao salvar as redes permitidas'
        }
      },
      vpn: {
        kvmUrl: 'Endereço do KVM',
        moreTip: 'Mais ações',
        updateTip: 'Atualizar para {{version}}',
        loading: 'Carregando...',
        okBtn: 'Sim',
        cancelBtn: 'Não',
        restart: 'Reiniciar o {{name}}?',
        update: 'Atualizar o {{name}} para {{version}}?',
        updateDesc: 'O daemon reinicia se estiver em execução. O login é mantido.',
        notInstall: 'O {{name}} não está instalado.',
        install: 'Instalar',
        installing: 'Instalando',
        installFailed: 'Falha na instalação',
        retry: 'Tentar novamente',
        connected: 'Conectado',
        connectedDesc:
          'Ligado entra na rede {{name}}. Desligado desconecta e para o {{name}} para liberar memória.',
        connectAtBoot: 'Conectar na inicialização',
        connectAtBootDesc: 'Entrar na rede {{name}} quando o KVM iniciar.',
        restartService: 'Reiniciar serviço',
        needsLogin: 'Requer login',
        error: 'Erro',
        thisDevice: 'Este dispositivo',
        memoryOf: '{{used}} (limite dos complementos {{limit}})',
        memoryPressed:
          'O grupo de memória dos complementos está perto do limite. Ativar a swap pode ajudar.',
        peersSummary: 'Peers: {{online}} online de {{total}}',
        peersSummaryIdle: 'Pares: {{online}} online, {{idle}} sob demanda, de {{total}}',
        showOffline: 'Mostrar offline ({{offline}})',
        hideOffline: 'Ocultar offline',
        blocked:
          '{{other}} está ligado ou conecta na inicialização. Só uma VPN pode funcionar por vez: desligue primeiro Conectado e Conectar na inicialização em {{other}}.',
        deviceName: 'Nome do dispositivo',
        deviceIP: 'IP do dispositivo',
        account: 'Conta',
        version: 'Versão',
        uptime: 'Tempo ativo',
        noPeers: 'Nenhum peer ainda.',
        online: 'Online',
        offline: 'Offline',
        idle: 'Sob demanda: o NetBird conecta quando o tráfego precisa',
        memory: 'Memória',
        uninstall: 'Desinstalar {{name}}',
        uninstallDesc:
          'Tem certeza de que deseja desinstalar o {{name}}? O login permanece na placa.',
        copy: 'Copiar',
        copied: 'Link copiado',
        copyFailed: 'Não foi possível copiar o link. Selecione-o e copie manualmente.',
        open: 'Abrir',
        checkAgain: 'Verificar novamente',
        notSignedIn: 'Ainda não conectado. Termine de entrar pelo link e verifique novamente.',
        checkFailed: 'Não foi possível verificar o status de login',
        loginWaiting: 'Esta página verifica a cada poucos segundos e continua quando você entrar.',
        uninstallFailed: 'Falha ao desinstalar',
        loginFailed: 'Falha no login'
      },
      tailscale: {
        title: 'Tailscale',
        download: 'Baixar o',
        package: 'pacote de instalação',
        unzip: 'e descompacte-o',
        notLogin:
          'O dispositivo ainda não foi vinculado. Por favor, faça login e vincule este dispositivo à sua conta.',
        urlPeriod: 'Esta URL é válida por 10 minutos',
        login: 'Login',
        logout: 'Sair',
        logoutDesc: 'Tem certeza de que deseja sair?',
        manualIntro: 'Ou instale manualmente via SSH:',
        copyBinaries: 'Copie tailscale e tailscaled para {{dir}} no IronKVM',
        linksFile: 'No mesmo diretório, crie um arquivo chamado links com estas duas linhas:',
        rebootRefresh: 'Reinicie o IronKVM e depois atualize esta página',
        loginServer: 'Servidor de controle',
        loginServerDesc:
          'Deixe vazio para usar o Tailscale. Digite a URL https de um servidor Headscale para usá-lo. Uma mudança inicia um novo login, e o servidor anterior continua se o novo não responder.',
        loginServerDefault: 'Tailscale (padrão)',
        loginServerInvalid: 'Digite uma URL https, por exemplo https://headscale.example.com',
        loginServerLocked: 'Saia da conta para mudar o servidor de controle.',
        loginServerSaved: 'Servidor de controle salvo'
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
        logout: 'Cancelar registro',
        logoutDesc:
          'Cancelar o registro remove este peer da sua conta NetBird e apaga a configuração dele aqui. Para entrar de novo é preciso uma chave de configuração ou um login SSO, e o peer pode receber um novo IP. Continuar?',
        joinFailed: 'Não foi possível entrar na rede'
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
            'O SHA-512 apenas verifica se o pacote corresponde ao manifesto fornecido por este servidor. Ele não comprova que o pacote seja uma versão oficial do IronKVM. Um servidor com falha ou mal-intencionado pode inutilizar o dispositivo, causar perda de dados ou comprometer o sistema.',
          confirm: 'Usar mesmo assim',
          useSipeed: 'Usar o servidor oficial da Sipeed',
          previewDisabled:
            'As atualizações de prévia ficam indisponíveis enquanto um servidor de atualização personalizado estiver ativado.'
        },
        offline: {
          chooseFile: 'Escolher arquivo',
          installing: 'Envio concluído. Instalando...',
          noFile: 'Nenhum arquivo escolhido',
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
          'O dispositivo instala a atualização e reinicia seu servidor. Esta página recarrega quando o servidor voltar.',
        releaseNotes: 'Notas de versão',
        ui: {
          title: 'Interface web',
          desc: 'Instale ou reverta a interface web sem atualizar o servidor',
          served: 'Em uso: {{version}}, API do servidor {{api}}',
          bundled: 'a interface incluída no aplicativo',
          bundledVersion: '{{version}} (incluída no aplicativo)',
          checkOff: 'A verificação de assinatura está desligada (/data/ironkvm/no-signature-check): qualquer pacote é instalado.',
          noKeys: 'O chaveiro está vazio: os pacotes são instalados sem verificação de assinatura.',
          check: 'Verificar',
          noPackage: 'O servidor de atualização não oferece nenhum pacote de interface web.',
          isCurrent: 'A interface web {{version}} já está em uso.',
          incompatible: 'A interface web {{version}} precisa da API do servidor {{api}}, e este servidor usa {{server}}. Atualize o aplicativo primeiro.',
          installTo: 'Instalar interface web {{version}}',
          installDesc: 'O dispositivo verifica a assinatura, muda para a nova interface e mantém a atual para reverter. Depois esta página é recarregada.',
          chooseFiles: 'Escolher pacote',
          chooseHint: 'ironkvm-ui_<version>.tar.gz e seu .minisig',
          noSignature: 'sem assinatura',
          pick: {
            noPackage: 'Escolha um pacote ironkvm-ui_<version>.tar.gz.',
            tooMany: 'Escolha um pacote e no máximo uma assinatura.',
            badName: 'Não é um pacote de interface web (ironkvm-ui_<version>.tar.gz).',
            signatureMismatch: 'A assinatura pertence a outro pacote.'
          },
          install: 'Instalar',
          installed: 'Interface web instalada. Recarregando...',
          rolledBack: 'Interface web revertida. Recarregando...',
          rollback: 'Reverter',
          rollbackDesc: 'Voltar para {{version}}',
          rollbackConfirm: 'Reverter a interface web para {{version}}?'
        }
      },
      account: {
        title: 'Conta',
        webAccount: 'Nome da Conta Web',
        role: 'Função',
        roles: { admin: 'Administrador', user: 'Usuário', viewer: 'Visualizador' },
        shareLinks: {
          title: 'Links de compartilhamento',
          desc: 'Um link de compartilhamento mostra a tela para quem o abrir. Ele não dá acesso a teclado, mouse, energia, mídia ou configurações e deixa de funcionar quando expira ou é revogado.',
          create: 'Criar link',
          label: 'Rótulo (opcional)',
          duration: 'Válido por',
          durations: {
            h1: '1 hora',
            h8: '8 horas',
            d1: '1 dia',
            d7: '7 dias',
            d30: '30 dias'
          },
          empty: 'Nenhum link de compartilhamento',
          expires: 'Expira em {{time}}',
          revoke: 'Revogar',
          revokeConfirm: 'Revogar este link? Quem estiver assistindo por ele será desconectado.',
          revoked: 'Link revogado',
          loadFailed: 'Falha ao carregar os links de compartilhamento',
          createFailed: 'Falha ao criar o link de compartilhamento',
          revokeFailed: 'Falha ao revogar o link de compartilhamento',
          createdTitle: 'Link de compartilhamento criado',
          createdDesc: 'Copie o link agora. Ele não poderá ser exibido novamente.'
        },
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
        mcpNote: 'Estas chaves não funcionam no MCP, que tem sua própria chave na página do MCP.',
        metricsUrl: 'URL das métricas',
        monitoring: 'Monitoramento',
        monitoringDesc:
          'O Prometheus lê as métricas com uma chave de API desta página, enviada como token Bearer. Qualquer função pode lê-las.',
        scrapeConfig: 'Configuração de scrape do Prometheus',
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
        kvmDescription: 'Opera o host remoto por meio de IronKVM.',
        switched: 'Função de agente trocada',
        switchFailed: 'Falha ao mudar de função de agente'
      },
      send: 'Enviar',
      stop: 'Parar',
      cancel: 'Cancelar',
      status: {
        connecting: 'Conectando ao gateway...',
        connected: 'Sessão PicoClaw conectada',
        disconnected: 'Sessão PicoClaw desconectada',
        stopped: 'Solicitação de parada enviada',
        stopFailed: 'Não foi possível enviar a solicitação de parada: o chat não está conectado.',
        stopTimeout: 'O PicoClaw não confirmou a parada. A tarefa ainda pode estar em execução.',
        runtimeStarted: 'Runtime do PicoClaw iniciado',
        runtimeStartFailed: 'Falha ao iniciar o runtime do PicoClaw',
        runtimeStopped: 'Runtime do PicoClaw interrompido',
        runtimeStopFailed: 'Falha ao parar o runtime do PicoClaw',
        controlSwitchedToMCP: 'Controle transferido para o serviço MCP externo'
      },
      connection: {
        runtime: {
          checking: 'Verificando',
          restoring: 'Restaurando o PicoClaw',
          ready: 'Runtime pronto',
          stopped: 'Runtime interrompido',
          blockedByMCP: 'O controle MCP externo está ativo',
          readyBlockedByMCP:
            'O runtime está em execução, mas um MCP externo controla agora a entrada do dispositivo.',
          readyWithoutControl:
            'O runtime está em execução. Conceda ao PicoClaw o controle do dispositivo antes de reconectar.',
          unavailable: 'Runtime indisponível',
          configError: 'Erro de configuração'
        },
        transport: {
          connecting: 'Conectando',
          connected: 'Conectado',
          disconnected: 'Desconectado',
          reconnect: 'Reconectar',
          reconnectDescription: 'Reconectar à sessão do PicoClaw em execução.',
          reconnectBlocked: 'O PicoClaw precisa do controle do dispositivo antes de reconectar.'
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
        picoclawDescription:
          'O PicoClaw pode enviar entrada de teclado e mouse. A entrada manual pode ser pausada.',
        mcp: 'Controle do dispositivo: MCP externo',
        mcpDescription:
          'O MCP externo pode escrever no dispositivo. O PicoClaw não assumirá a entrada.',
        off: 'Controle do dispositivo: desativado',
        offDescription:
          'A IA não enviará entrada de teclado ou mouse. O controle manual continua disponível.',
        transitioning: 'Controle do dispositivo: alternando',
        transitioningDescription: 'O controle do dispositivo está sincronizando. Aguarde.',
        grant: 'Conceder controle',
        release: 'Liberar',
        releasing: 'Liberando...',
        switching: 'Alternando...',
        releasingLabel: 'Controle do dispositivo: liberando',
        releasingDescription:
          'O controle do dispositivo está sendo devolvido. O PicoClaw parou as escritas em andamento.',
        granted: 'Controle do PicoClaw concedido',
        released: 'Controle do PicoClaw liberado',
        grantFailed: 'Falha ao conceder controle ao PicoClaw',
        releaseFailed: 'Falha ao liberar controle do PicoClaw',
        grantConfirmTitle: 'Alternar controle do dispositivo para PicoClaw?',
        grantConfirmDesc: 'As gravações de dispositivo do MCP externo serão interrompidas.'
      },
      install: {
        install: 'Instalar PicoClaw',
        name: 'PicoClaw (build do IronKVM)',
        update: 'Atualizar para {{version}}',
        installing: 'Instalando PicoClaw',
        success: 'PicoClaw instalado com sucesso',
        failed: 'Falha ao instalar PicoClaw',
        uninstalling: 'Desinstalando o runtime...',
        uninstalled: 'Runtime desinstalado com sucesso.',
        uninstallFailed: 'Falha na desinstalação.',
        requiredTitle: 'PicoClaw não está instalado',
        requiredDescription: 'Instale o PicoClaw antes de iniciar o runtime do PicoClaw.',
        progressDescription: 'PicoClaw está sendo instalado.',
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
        apiKeyOptionalPlaceholder: 'Não é necessária para ollama, lmstudio ou vllm',
        save: 'Salvar',
        saving: 'Salvando',
        saved: 'Configuração do modelo salva',
        saveFailed: 'Falha ao salvar a configuração do modelo',
        invalid: 'Identificador do modelo, API Base URL e chave API são obrigatórios',
        invalidNoKey: 'Identificador do modelo e API Base URL são obrigatórios'
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
        switchFromMCP: 'Mudar para o PicoClaw e iniciar',
        takeoverAndStart: 'Assumir e iniciar'
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
    speaker: {
      title: 'Alto-falante',
      unmute: 'Ativar som',
      mute: 'Silenciar',
      hostIdle: 'O host não está enviando áudio',
      hostIdleHint: 'Reproduza algo no host ou escolha o KVM como saída de som.',
      room: 'Microfone do ambiente',
      roomListen: 'Ouvir o ambiente',
      roomVolume: 'Volume',
      roomHint: 'Todos que assistem veem quando está ligado.',
      roomLive: 'Microfone ativo',
      roomLiveBy: 'Microfone do ambiente ligado: {{names}}'
    },
    upstream: {
      check: 'Verificar atualizações',
      updateTo: 'Atualizar para {{version}}',
      confirm: 'Atualizar {{name}} para {{version}}?',
      confirmDesc:
        'A nova versão é baixada do GitHub e verificada com os checksums que ela publica. Se algo falhar, a versão atual permanece.',
      ok: 'Atualizar',
      upToDate: 'Atualizado',
      builtIn: 'embutida',
      checkFailed: 'Não foi possível verificar atualizações: {{error}}',
      unverifiable: 'A versão {{version}} não é oferecida: {{reason}}',
      inUse: 'Não é possível atualizar agora: {{reason}}',
      running: 'Atualizando para {{version}}...',
      done: '{{name}} atualizado para {{version}}',
      failed: 'A última atualização falhou: {{error}}'
    },
    menu: {
      mediaAdd: 'Adicionar imagem',
      mediaMoreOptions: 'Mais opções',
      mediaSettings: 'Configurações de mídia',
      collapse: 'Recolher Menu',
      expand: 'Expandir Menu',
      more: 'Mais',
      media: 'Mídia',
      tools: 'Ferramentas',
      text: 'Texto',
      advanced: 'Avançado',
      mediaMounted: 'Montado',
      mediaLibrary: 'Biblioteca',
      textToHost: 'Para o host',
      textFromHost: 'Do host'
    },
    serverContract: {
      incompatibleTitle: 'Esta interface web não suporta este servidor',
      incompatible: 'A interface web suporta a API do servidor {{ui}}, e o servidor usa a API {{server}}. Algumas funções podem falhar. Atualize o aplicativo ou reverta a interface web em Configurações > Atualização.',
      dismiss: 'Fechar'
    },
    ion: {
      checking: 'Verificando a memória de vídeo antes de iniciar a transmissão...',
      warn: 'A memória de vídeo está baixa. Uma única reinicialização do servidor a esgotaria. Reinicie quando for conveniente.',
      criticalTitle: 'Memória de vídeo insuficiente para iniciar a transmissão',
      criticalBody:
        'Iniciar o vídeo esgotaria a memória reservada e pararia o servidor. Todas as outras funções continuam funcionando, incluindo o controle de energia e a reinicialização. Só uma reinicialização do IronKVM recupera essa memória.',
      criticalContinue: 'Iniciar o vídeo mesmo assim',
      criticalReboot: 'Reiniciar o IronKVM',
      criticalRebooting: 'Reiniciando...'
    },
    failsafe: {
      title: 'Vídeo desligado: o servidor iniciou em modo de segurança',
      body: 'A biblioteca de vídeo derrubou o servidor várias vezes seguidas, então o servidor agora funciona sem vídeo. Todo o resto funciona: configurações, atualizações, logs, controle de energia e terminal. Tente o vídeo novamente após uma atualização ou correção. Reiniciar o IronKVM também tenta o vídeo novamente.',
      showLog: 'Mostrar log',
      hideLog: 'Ocultar log',
      retry: 'Tentar vídeo novamente',
      retrying: 'Reiniciando...'
    },
    totp: {
      login: {
        title: 'Autenticação de dois fatores',
        description:
          'Digite o código de 6 dígitos do seu aplicativo autenticador, ou um dos seus códigos de recuperação.',
        placeholder: 'Código ou código de recuperação',
        noEmptyCode: 'Digite um código',
        verify: 'Verificar',
        back: 'Voltar para a senha',
        invalidCode: 'Código inválido',
        locked: 'Muitos códigos errados. Tente novamente em alguns minutos.',
        expired: 'O login expirou. Digite sua senha novamente.'
      },
      settings: {
        title: 'Autenticação de dois fatores',
        description: 'Pedir um código de um aplicativo autenticador depois da senha.',
        scope:
          'Vale apenas para o login web. VNC, IPMI e Redfish continuam usando só a senha, e as chaves de API funcionam sem código.',
        on: 'Ativada',
        off: 'Desativada',
        enable: 'Ativar',
        disable: 'Desativar',
        regenerate: 'Novos códigos de recuperação',
        codesLeft: 'Códigos de recuperação restantes: {{count}}',
        enableTitle: 'Ativar autenticação de dois fatores',
        disableTitle: 'Desativar autenticação de dois fatores',
        regenerateTitle: 'Novos códigos de recuperação',
        confirmPassword: 'Confirme com sua senha',
        continue: 'Continuar',
        scan: 'Escaneie o código QR com um aplicativo autenticador, ou digite o segredo manualmente. Depois digite o código que o aplicativo mostra.',
        secret: 'Segredo',
        enterCode: 'Código do aplicativo',
        verifyAndEnable: 'Verificar e ativar',
        codesOnce:
          'Salve estes códigos de recuperação agora. Cada um funciona uma vez no lugar de um código, e eles não serão mostrados novamente.',
        recoveryCodes: 'Códigos de recuperação do IronKVM',
        download: 'Baixar',
        saved: 'Já salvei',
        disabled: 'Autenticação de dois fatores desativada',
        wrongPassword: 'A senha está incorreta',
        setupExpired: 'A configuração expirou. Comece novamente.',
        loadFailed: 'Falha ao carregar o estado da autenticação de dois fatores',
        failed: 'A alteração falhou'
      },
      users: {
        tag: '2FA',
        reset: 'Redefinir 2FA',
        resetConfirm: 'Desativar a autenticação de dois fatores de {{username}}?',
        resetDone: 'Autenticação de dois fatores redefinida'
      }
    },
    tlsCert: {
      title: 'Certificado',
      description: 'O certificado apresentado pelo servidor HTTPS.',
      uploaded: 'Enviado',
      selfSigned: 'Autoassinado',
      httpsOff: 'O HTTPS está desativado. O certificado será usado quando o HTTPS for ativado.',
      none: 'Nenhum certificado ainda. Um será gerado quando o HTTPS for ativado.',
      subject: 'Assunto',
      issuer: 'Emissor',
      names: 'Nomes',
      validFrom: 'Válido desde',
      validUntil: 'Válido até',
      keyType: 'Chave',
      chain: 'Certificados na cadeia',
      fingerprint: 'SHA-256',
      restore: 'Restaurar autoassinado',
      restoreConfirm: 'Voltar ao certificado autoassinado?',
      restoreDesc:
        'O certificado e a chave enviados serão excluídos. Os navegadores voltarão a avisar sobre o certificado autoassinado.',
      restored: 'Certificado autoassinado restaurado',
      restoreFailed: 'Falha ao restaurar o certificado autoassinado',
      uploadTitle: 'Enviar um certificado',
      uploadDescription:
        'Formato PEM. Coloque primeiro o certificado do servidor, depois os intermediários. A chave não pode estar criptografada.',
      certField: 'Cadeia de certificados',
      keyField: 'Chave privada',
      chooseFile: 'Escolher arquivo',
      readFailed: 'Não foi possível ler o arquivo',
      check: 'Verificar',
      install: 'Instalar',
      installConfirm: 'Instalar este certificado?',
      installDesc:
        'Novas conexões passam a usá-lo imediatamente. As páginas abertas continuam funcionando; recarregue para ver o novo certificado.',
      installed: 'Certificado instalado',
      installFailed: 'Falha ao instalar o certificado',
      checkFailed: 'O certificado foi recusado',
      loadFailed: 'Falha ao carregar o certificado',
      warnings: {
        expiresSoon: 'O certificado expira em menos de 30 dias.',
        nameMismatch:
          'O certificado não cobre: {{names}}. Os navegadores vão avisar quando o dispositivo for aberto por esses nomes.',
        selfSigned:
          'Este certificado é autoassinado. Os navegadores só confiam nele depois que é instalado em cada computador.'
      },
      errors: {
        tooLarge: 'O arquivo é grande demais.',
        noCertificate: 'Nenhum certificado PEM encontrado.',
        badCertificate:
          'Não é possível ler o certificado. Cole um texto PEM que comece com -----BEGIN CERTIFICATE-----.',
        noKey: 'Nenhuma chave privada PEM encontrada.',
        encryptedKey: 'A chave privada está criptografada. Remova a senha da chave primeiro.',
        badKey: 'Não é possível ler a chave privada.',
        keyMismatch: 'A chave privada não pertence ao certificado.',
        expired: 'O certificado, ou um da sua cadeia, expirou.',
        notYetValid: 'O certificado ainda não é válido. Verifique o relógio do dispositivo.',
        badChain:
          'A cadeia está fora de ordem ou incompleta. Coloque primeiro o certificado do servidor, depois cada emissor.',
        weakKey: 'A chave é curta demais. Use RSA 2048 ou maior, ou ECDSA.',
        notServerCert: 'O certificado não é para um servidor TLS.'
      }
    }
  }
};

export default pt_br;
