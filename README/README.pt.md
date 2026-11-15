**Languages:** [English](../README.md) · [Polski](README.pl.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Italiano](README.it.md) · [Português](README.pt.md) · [Nederlands](README.nl.md) · [Svenska](README.sv.md) · [Norsk](README.no.md) · [Dansk](README.da.md) · [Suomi](README.fi.md) · [Čeština](README.cs.md) · [Slovenčina](README.sk.md) · [Magyar](README.hu.md) · [Română](README.ro.md) · [Türkçe](README.tr.md) · [Bahasa Indonesia](README.id.md) · [Bahasa Melayu](README.ms.md) · [Tiếng Việt](README.vi.md) · [Русский](README.ru.md) · [Українська](README.uk.md) · [Български](README.bg.md) · [Ελληνικά](README.el.md) · [العربية](README.ar.md) · [עברית](README.he.md) · [हिन्दी](README.hi.md) · [中文](README.zh.md) · [日本語](README.ja.md) · [한국어](README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](../LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](../go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#download-e-execução)

Uma pequena ferramenta portátil para enviar arquivos e pastas diretamente entre dois computadores na mesma rede local. Sem nuvem, sem conta e sem servidor intermediário: os dados vão diretamente de um PC para o outro através de uma conexão criptografada. Núcleo em Go, uma janela WebView2 e uma versão de linha de comando para scripts.

![Janela principal do LAN Transfer: um receptor emparelhado e três itens prontos para envio](../docs/screenshots/main.png)

## Conteúdo

- [Funcionalidades](#funcionalidades)
- [Idioma da interface](#idioma-da-interface)
- [Capturas de tela](#capturas-de-tela)
- [Download e execução](#download-e-execução)
- [Como emparelhar](#como-emparelhar)
- [Envio de arquivos](#envio-de-arquivos)
- [Linha de comando](#linha-de-comando)
- [Portas e firewall](#portas-e-firewall)
- [Notas de segurança](#notas-de-segurança)
- [Variáveis de ambiente](#variáveis-de-ambiente)
- [Compilação e testes](#compilação-e-testes)
- [Licença](#licença)

## Funcionalidades

- Descoberta automática de outros computadores que executam o LAN Transfer (broadcast UDP, com uma varredura limitada da rede /24 como alternativa).
- Emparelhamento com um código de 6 dígitos exibido no receptor. Sem emparelhamento, nada pode ser enviado e nenhuma pasta pode ser navegada.
- Modo opcional sem código: no recetor, desative **Exigir código de emparelhamento** nas Definições (guardado em `data\settings.json`) ou inicie `serve --no-code`. Assim qualquer computador com o LAN Transfer na rede pode enviar-lhe ficheiros sem código, sempre cifrados; entretanto o cabeçalho mostra **Sem código**.
- Criptografia TLS 1.3. Cada execução cria um novo certificado autoassinado; sua impressão digital é fixada durante o emparelhamento e verificada em cada conexão.
- Arquivos e pastas inteiras, com a estrutura de pastas preservada e até 4 arquivos enviados em paralelo.
- Retomada de transferências interrompidas e verificação SHA-256 de cada arquivo.
- Os dados incompletos ficam em uma pasta oculta `.lantransfer-partial`, de modo que um arquivo truncado nunca aparece com seu nome final.
- Conflitos de nomes geram `file (1).txt` ou `Folder (1)`; arquivos existentes nunca são sobrescritos.
- Nomes perigosos são rejeitados (`..`, caminhos absolutos, fluxos de dados alternativos `file:stream`, nomes reservados como `CON`, `NUL`, `COM1`). O espaço livre em disco é verificado antes do recebimento.
- Arrastar e soltar a partir do Explorador de Arquivos, uma fila com progresso por arquivo, velocidade, tempo restante, cancelamento, retomada e um histórico de itens enviados e recebidos.
- Portátil: um único EXE, com configurações e histórico em uma pasta `data` ao lado dele.
- A janela, o título dela e as mensagens da linha de comando em 30 idiomas (os mesmos deste README), escolhidos automaticamente de acordo com o idioma de exibição do Windows; layout da direita para a esquerda para árabe e hebraico.

## Idioma da interface

O LAN Transfer fala os mesmos 30 idiomas deste README. Na primeira execução, ele usa o idioma de exibição do Windows (o primeiro idioma compatível da lista de idiomas do Windows); se o Windows usar um idioma que não está na lista, a interface fica em inglês.

Para mudá-lo, clique no botão de engrenagem (**Definições**) no canto superior direito e escolha um idioma. A janela, o título dela e todas as mensagens mudam imediatamente, e a escolha é salva em `data\settings.json`. **Automático (Windows)** volta a seguir o idioma de exibição do Windows. Números, tamanhos, velocidades, datas e formas de plural seguem as regras do idioma escolhido; árabe e hebraico usam um layout espelhado, da direita para a esquerda, enquanto caminhos, endereços IP, códigos e impressões digitais continuam escritos da esquerda para a direita. As mensagens vindas do outro computador (por exemplo, “espaço insuficiente no disco do destinatário”) são exibidas no seu idioma, não no do receptor.

O `LANTransferCLI.exe` usa o mesmo idioma. `LANTRANSFER_LANG` (por exemplo, `LANTRANSFER_LANG=de`) tem prioridade sobre a escolha salva, tanto na janela quanto na linha de comando.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![LAN Transfer em inglês](../docs/screenshots/lang-en.png) | ![LAN Transfer em alemão](../docs/screenshots/lang-de.png) | ![LAN Transfer em árabe, layout da direita para a esquerda](../docs/screenshots/lang-ar.png) |

## Capturas de tela

| Emparelhamento | Envio com progresso |
|---|---|
| ![Diálogo de emparelhamento com o código de 6 dígitos e a impressão digital do certificado](../docs/screenshots/pairing.png) | ![Fila de transferências com progresso, velocidade e tempo restante](../docs/screenshots/sending.png) |
| **Retomada após cancelamento** | **Receptor: histórico** |
| ![Uma transferência cancelada retomada a partir do ponto já recebido, verificada com SHA-256](../docs/screenshots/resume.png) | ![Janela do receptor com os itens recebidos no histórico](../docs/screenshots/history.png) |

## Download e execução

1. Baixe `LANTransfer-2.1.0-windows-amd64.zip` em [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) e verifique-o com o arquivo `.sha256` publicado.
2. Extraia-o para qualquer pasta em ambos os computadores.
3. Inicie `LANTransfer.exe`. Na primeira execução, o Windows pede permissão (UAC) para adicionar duas regras de firewall; consulte [Portas e firewall](#portas-e-firewall).

Requisitos: Windows 10 ou 11 (64 bits) com o runtime do WebView2, que faz parte das versões atuais do Windows. Ambos os computadores devem executar a versão 2. A versão 2 não consegue se comunicar com a versão 1.

## Como emparelhar

1. Inicie o LAN Transfer em ambos os computadores. O receptor mostra seu código de emparelhamento e a impressão digital do certificado no canto superior direito.
2. No computador remetente, clique no receptor na lista **Destinatário**. Se ele não aparecer, use **Ligar manualmente (endereço IP)** e digite o endereço IP dele.
3. Digite o código de 6 dígitos exibido na tela do receptor e compare a impressão digital do certificado mostrada no diálogo com a do receptor. Se forem diferentes, não continue.
4. Clique em **Emparelhar**. O dispositivo passa a mostrar **✓ emparelhado**.

Após 5 códigos incorretos vindos de um mesmo computador, o receptor bloqueia novas tentativas desse computador. **Novo código** cria um novo código e cancela todos os emparelhamentos existentes. O receptor cria um novo código e um novo certificado sempre que é iniciado; portanto, depois que o receptor for reiniciado, emparelhe novamente.

## Envio de arquivos

1. Emparelhe com o receptor.
2. Arraste arquivos ou pastas para a janela, use **Selecionar ficheiros** / **Selecionar pasta** ou cole um caminho.
3. Escolha o destino: a pasta padrão do receptor (`Downloads\LAN Transfer`) ou outra pasta existente no receptor.
4. Clique em **Enviar**. Cada item mostra progresso, velocidade e tempo restante, e pode ser cancelado. Uma transferência cancelada ou interrompida pode ser retomada; apenas a parte que falta é enviada.

## Linha de comando

`LANTransferCLI.exe` é o mesmo programa sem janela, para scripts e automação.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

O código também pode ser passado em `LANTRANSFER_CODE`. Um receptor sem janela pode usar um código fixo definido em `LANTRANSFER_PIN`. Executar novamente o mesmo comando `send` retoma uma transferência interrompida.

## Portas e firewall

- UDP 45454: descoberta. TCP 45455: transferência (HTTPS).
- Na primeira execução, o aplicativo pede uma única vez (UAC) para adicionar duas regras de entrada: `LAN Transfer 2 Discovery (UDP)` e `LAN Transfer 2 Transfer (TCP)`.
- As regras aceitam conexões apenas da sub-rede local (`remoteip=LocalSubnet`) e apenas nos perfis de rede Privada e Domínio (`profile=private,domain`). Em uma rede que o Windows marca como Pública (café, hotel, aeroporto), outros computadores não conseguem se conectar ao LAN Transfer.
- Se o receptor não estiver visível na sua rede doméstica, verifique nas configurações de rede do Windows se a rede está definida como Privada.
- A versão 1 criava regras chamadas `LAN Transfer Discovery` e `LAN Transfer Transfer` para todos os perfis, incluindo o Público. Se a versão 2 as encontrar, exibe um banner oferecendo substituí-las. Nada é alterado até que você clique em **Atualizar regras** e confirme o aviso do Windows.

## Notas de segurança

- A descoberta via UDP não é autenticada; serve apenas como indicação. A confiança vem do emparelhamento: o código mais a impressão digital do certificado fixada.
- Durante o primeiro emparelhamento, um atacante ativo na mesma rede (man in the middle) poderia se passar pelo receptor. Compare a impressão digital do certificado em ambas as telas antes de emparelhar.
- Um remetente emparelhado pode navegar pelas pastas do receptor e salvar arquivos em qualquer pasta existente nele (a opção de destino personalizado). Emparelhe apenas computadores em que você confia e use **Novo código** para revogar todos os emparelhamentos.
- As regras de firewall não abrem as portas em redes marcadas como Públicas.
- Com **Exigir código de emparelhamento** desativado, qualquer pessoa na rede que use o LAN Transfer pode enviar ficheiros para este computador e explorar as suas pastas sem código, como na versão 1 (mas cifrado). Use-o apenas numa rede de confiança. Voltar a ativar o código desliga os remetentes ligados sem ele.

## Variáveis de ambiente

Destinam-se a testes e automação; o uso normal não requer nenhuma delas.

| Variável | Significado |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | outras portas |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | pasta de recebimento, pasta de dados do programa |
| `LANTRANSFER_NAME` | nome do computador exibido aos outros |
| `LANTRANSFER_LANG` | idioma da interface e da linha de comando (`en`, `de`, `pl`, …); tem prioridade sobre a escolha salva |
| `LANTRANSFER_LOOPBACK=1` | apenas 127.0.0.1: sem broadcasts, sem varredura da LAN, sem regras de firewall |
| `LANTRANSFER_PEER_UDP_PORTS` | portas UDP de outras instâncias no modo loopback |
| `LANTRANSFER_NO_FIREWALL=1` | não criar nem verificar regras de firewall |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | código fixo do receptor, código usado pelo remetente |
| `LANTRANSFER_REQUIRE_CODE=0` | `0` = este recetor não exige código de emparelhamento, `1` = código obrigatório; prevalece sobre a definição |

## Compilação e testes

Go 1.23 ou mais recente no Windows 10/11.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` executa os testes e compila `LANTransfer.exe` (janela) e `LANTransferCLI.exe` (console) com a versão de `VERSION` (padrão `2.1.0`). Os testes incluem transferências de ponta a ponta via 127.0.0.1: emparelhamento e bloqueio, rejeição de requisições não autenticadas, arquivos de 0 B a 96 MB, pastas Unicode e caminhos com mais de 260 caracteres, conflitos de nomes, nomes maliciosos, uma soma de verificação incorreta, cancelamento e retomada, um disco cheio e fixação de certificado.

As traduções ficam em `locales/ui.<code>.json` (o polonês é o idioma de origem e o inglês, o de reserva). `tools/check_i18n.py` verifica se os 30 arquivos têm as mesmas chaves, marcadores de posição e formas de plural e se não contêm restos de polonês; um teste em Go verifica se cada chave usada pelo programa existe em todos os idiomas.

Notas de versão: [docs/releases/v2.1.0.md](../docs/releases/v2.1.0.md) · [v2.0.0](../docs/releases/v2.0.0.md).

## Licença

MIT, consulte [LICENSE](../LICENSE).
