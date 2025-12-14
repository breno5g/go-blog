---
title: "Plugin Para O KoReader"
date: 2025-12-14T18:00:34-03:00
draft: false
description: "Acompanhe a criação de um plugin para o KoReader"
tags: ["books", "code"]
categories: []
---

# O problema

Recentemente decidi retomar meus estudos, tanto para evoluir na programação quanto para continuar meu desenvolvimento pessoal.  
Alguns amigos me recomendaram livros de psicologia, filosofia e outros temas que despertaram meu interesse — e, claro, fui adicionando tudo ao meu Kindle.

Minha ideia era simples: ler no Kindle, marcar os trechos importantes e enviar essas anotações automaticamente para meu **vault do Obsidian**, onde centralizo meus estudos e notas. Assim eu poderia revisitar facilmente os conceitos mais marcantes de cada leitura.

Mas aí veio o problema.  
Meu plano era usar o **Readwise** para capturar os highlights e enviá-los para o Obsidian através de um plugin simples.  
Para minha ingrata surpresa, descobri que o Readwise **não captura os destaques de livros adicionados manualmente ao Kindle** — apenas os comprados na loja da Amazon.

E eu sabia que os highlights existiam, pois apareciam normalmente ao abrir o livro no aplicativo oficial.

# A solução

Eu poderia simplesmente desistir da automação e continuar copiando tudo manualmente.  
Ou poderia procurar alguma ferramenta de terceiros que resolvesse o problema.  
Mas decidi seguir outro caminho: **abandonar de vez o ecossistema fechado da Amazon**, desbloquear meu Kindle e criar minha própria solução.

Depois de algumas pesquisas sobre jailbreak, acabei encontrando um artigo sobre o **KOReader** — um leitor de documentos _open source_ para dispositivos _e-ink_, com suporte a plugins e ampla personalização. Exatamente o que eu procurava.

Libertada a alma Kindle, instalei o KOReader e parti em busca de documentação sobre como desenvolver plugins.  
Mas, para minha surpresa, **não havia nenhum “getting started”** — nem mesmo uma barra de pesquisa para localizar funções na wiki.

Ou seja: se eu quisesse fazer isso funcionar, teria que explorar o código-fonte e aprender como os antigos faziam.

# O plano de ação

Antes de partir para o estudo do código, resolvi definir os requisitos básicos de como esse plugin deve se comportar.

A ideia básica é que o plugin deve:

- Buscar os **highlights** do livro atualmente aberto;
- Ler o conteúdo salvo;
- Enviar para um **servidor HTTP** que deve lidar com isso de forma apropriada.

Assim o plugin poderá se comunicar com qualquer servidor que tenha uma rota para lidar com isso.

Para a parte visual, precisamos de:

- Um **formulário** onde o usuário possa inserir o IP e a porta onde o servidor estará rodando;
- Um **botão** para disparar o evento de sincronização;
- E, claro, um **botão de debug**.

# Botando a mão na massa

Como ponto de partida, escolhi dois plugins nativos do KOReader como base:

- O do _Calibre_, que me daria uma base de como exibir os componentes em tela;
- E o _Exporter_, que é um plugin já pensado para lidar com os highlights.

A primeira tarefa a ser feita era descobrir como iniciar o plugin e mostrar um singelo “Hello world” em tela — e, para isso, a estrutura abaixo foi necessária:

```lua
local InfoMessage = require("ui/widget/infomessage")
local UIManager = require("ui/uimanager")
local WidgetContainer = require("ui/widget/container/widgetcontainer")
local _ = require("gettext")

local ObsidianSync = WidgetContainer:extend({
	name = "obsidiansync",
	is_doc_only = false,
})

function ObsidianSync:init()
	self.ui.menu:registerToMainMenu(self)
end

function ObsidianSync:addToMainMenu(menu_items)
	menu_items.obsidiansync = {
		text = _("Obsidian Sync"),
		sorting_hint = "tools",
		sub_item_table = {
			{
				text = _("Export Highlights"),
				callback = function()
					UIManager:show(InfoMessage:new({
						text = _("Hello, plugin world"),
					}))
				end,
			},
		},
	}
end

return ObsidianSync
```

Explicando por alto esse código:  
fizemos a inicialização de uma classe chamada **`ObsidianSync`**, que herda da classe **`WidgetContainer`** todos os métodos e atributos básicos para podermos criar um _plugin/widget_ no KOReader.

Criamos também um método **`init`**, que será chamado na inicialização do app e será responsável por adicionar o plugin na interface.

Para finalizar, criamos o método **`addToMainMenu`** que, como o nome sugere, fica responsável por adicionar ao menu os elementos visuais do plugin.

## Evento para capturar informações basicas
Após criarmos a estrutura básica do plugin, eu precisava descobrir como capturar as informações básicas do livro que estamos mexendo, coisas como título, quantidade de páginas, progresso de leitura e, não menos importante, os **highlights**.

​Então, para começarmos essa parte, resolvi criar um método de **`debug`** que nos retornasse as informações básicas de qual livro estamos, qual a extensão dele e em qual diretório ele está localizado.

```lua
function ObsidianSync:getFileNameAndExtension(path)
	local info = {
		dirname = path:match("(.+)/[^/]+$"),
		filename = path:match("([^/]+)%."),
		extension = path:match("%.([^%.]+)$"),
	}

	return info
end

function ObsidianSync:debug()
	local info = {}

	table.insert(info, "====== DEBUG INFO ======")
	table.insert(info, "")

	if not self.ui.document then
		table.insert(info, "❌ Open a document to proced")
		self:info(table.concat(info, "\n"))
		return
	end

	local fileInfo = self:getFileNameAndExtension(self.ui.document.file)
	table.insert(info, "✅ Document infos:")
	table.insert(info, "- Dirname: " .. fileInfo.dirname)
	table.insert(info, "- Filename: " .. fileInfo.filename)
	table.insert(info, "- Extension: " .. fileInfo.extension)

	self:info(table.concat(info, "\n"))
end
```

Provavelmente, temos alguma abordagem mais simples para pegar os dados do livro, mas nada que um `regex` e certa força de vontade não resolvam.
## Acessando os highlights
Agora que temos acesso às informações mais básicas do livro atual, precisamos descobrir onde o KoReader salva as informações que queremos do livro.

​Depois de alguns minutos de pesquisa e testes, acabei vendo que na mesma pasta onde o livro está salvo é gerada uma outra pasta chamada *filename.sdr*  e, nessa pasta, é criado um arquivo *metadata.{extensão do livro}.lua* que contém todas as informações básicas sobre o livro.

​Analisando o arquivo metadata, encontrei a chave *annotation* que contém todos os **highlights** e informações dele salvos.

```lua
function ObsidianSync:getSDRData()
	if not self.ui.document then
		self:info("❌ Open a document to proced")
		return
	end

	local fileInfo = self:getFileNameAndExtension()
	local chunk, err =
		loadfile(fileInfo.dirname .. "/" .. fileInfo.filename .. ".sdr/metadata." .. fileInfo.extension .. ".lua")
	if not chunk then
		self:info("❌ Error to open sdr: " .. err)
		return
	end

	local metadata = chunk()
	self:info(metadata["annotations"][1]["text"])
end
```

## Projetando a comunicação com o servidor
Agora que temos os dados que precisamos, falta apenas conseguirmos nos comunicar com o servidor e enviar os arquivos.

​Pensei em N formas que poderíamos fazer essa conexão, mas resolvi focar primeiro no básico e depois incrementar, caso necessário. Então, para começar, precisamos apenas do **`IP`** da máquina onde o servidor estará rodando e da **`PORTA`** onde o serviço vai estar escutando.

​Antes de tentar a comunicação com o servidor, resolvi criar um arquivo de configuração básico para que o usuário possa preencher essas informações uma única vez e possa reutilizar nas próximas sessões de leitura.

```lua
local ObsidianSync = WidgetContainer:extend({
	name = "obsidiansync",
	is_doc_only = false,

	defaults = {
		address = "127.0.0.1",
		port = 9090,
		password = "",
	},
})

function ObsidianSync:configure(touchmenu_instance)
	local MultiInputDialog = require("ui/widget/multiinputdialog")
	local url_dialog

	local current_settings = self.settings
		or G_reader_settings:readSetting("obsidiansync_settings", ObsidianSync.defaults)

	local obsidian_url_address = current_settings.address
	local obsidian_url_port = current_settings.port

	url_dialog = MultiInputDialog:new({
		title = _("Set custom obsidian address"),
		fields = {
			{
				text = obsidian_url_address,
				input_type = "string",
				hint = _("IP Address"),
			},
			{
				text = tostring(obsidian_url_port),
				input_type = "number",
				hint = _("Port"),
			},
		},
		buttons = {
			{
				{
					text = _("Cancel"),
					id = "close",
					callback = function()
						UIManager:close(url_dialog)
					end,
				},
				{
					text = _("OK"),
					callback = function()
						local fields = url_dialog:getFields()
						if fields[1] ~= "" then
							local port = tonumber(fields[2])
							if not port or port < 1 or port > 65355 then
								port = ObsidianSync.defaults.port
							end

							-- Preserva o device_id existente ao salvar
							local new_settings = {
								address = fields[1],
								port = port,
								device_id = self.settings.device_id,
							}
							G_reader_settings:saveSetting("obsidiansync_settings", new_settings)
							self.settings = new_settings
							self:showNotification(_("✅ Settings saved!"))
						end
						UIManager:close(url_dialog)
						if touchmenu_instance then
							touchmenu_instance:updateItems()
						end
					end,
				},
			},
		},
	})
	UIManager:show(url_dialog)
	url_dialog:onShowKeyboard()
end
```

Esse método fica responsável por criar os inputs e salvar os dados preenchidos pelo usuário num arquivo de configuração que permanece independente do livro ou sessão atual.

​Agora que temos os dados de para onde enviar os arquivos, podemos apenas pegar o metadata e enviá-lo para o servidor — o metadata é basicamente um objeto (table) e iremos deixar o servidor lidar com os dados que ele quer ou não.

```lua
function ObsidianSync:sendToServer()
	local json = require("json")

	local metadata = self:getSDRData()
	if not metadata then
		return
	end

	local body, err = json.encode(metadata)
	if not body then
		self:showNotification(_("❌ Error encoding JSON: ") .. (err or _("unknown")), 5)
		return
	end

	local device_id = self:getDeviceID()
	if not device_id or device_id == "" then
		self:showNotification(_("❌ Error: Could not get device ID."), 5)
		return
	end

	local settings = self.settings or G_reader_settings:readSetting("obsidiansync_settings", ObsidianSync.defaults)
	local url = "http://" .. settings.address .. ":" .. settings.port .. "/sync"

	self:showNotification(_("Syncing with server..."), 2)

	UIManager:scheduleIn(0.25, function()
		self:_doSyncRequest(url, body, device_id)
	end)
end
```

Aqui pegamos todo o _metadata_ e, com as configurações salvas do usuário, nos comunicamos com o servidor na rota **`/sync`** para finalizarmos a **sincronização dos** _**highlights**_.
# Conclusão
Chegamos ao final do desenvolvimento do _plugin_ (ao menos a parte do KoReader) e o código completo está disponível no meu _github_ no repositório [koreadersync.koplugin](https://github.com/breno5g/koreadersync.koplugin).