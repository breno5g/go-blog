---
title: "Plugin For KoReader"
date: 2025-12-14T18:00:34-03:00
draft: false
description: "Follow along as I build a plugin for KoReader"
tags: ["books", "code"]
categories: []
---

# The problem

I recently decided to get back to studying, both to improve my programming skills and to continue my personal development.  
Some friends recommended books on psychology, philosophy, and other topics that caught my interest — and, of course, I kept adding everything to my Kindle.

My idea was simple: read on the Kindle, highlight the important parts, and automatically send those annotations to my **Obsidian vault**, where I keep all my studies and notes. That way I could easily revisit the most striking concepts from each book.

But then came the problem.  
My plan was to use **Readwise** to capture the highlights and send them to Obsidian through a simple plugin.  
To my unpleasant surprise, I discovered that Readwise **doesn't capture highlights from books manually added to the Kindle** — only those purchased from the Amazon store.

And I knew the highlights existed, because they showed up normally when I opened the book in the official app.

# The solution

I could have just given up on automation and kept copying everything manually.  
Or I could have looked for some third-party tool to solve the problem.  
But I decided to go another route: **completely abandon Amazon's closed ecosystem**, jailbreak my Kindle, and create my own solution.

After some research on jailbreaking, I ended up finding an article about **KOReader** — an _open source_ document reader for _e-ink_ devices, with plugin support and extensive customization. Exactly what I was looking for.

With the Kindle soul freed, I installed KOReader and started looking for documentation on how to develop plugins.  
But, to my surprise, **there was no "getting started"** — not even a search bar to find functions in the wiki.

In other words: if I wanted to make this work, I'd have to explore the source code and learn how the ancients did it.

# The action plan

Before diving into the code, I decided to define the basic requirements for how this plugin should behave.

The basic idea is that the plugin should:

- Fetch the **highlights** from the currently open book;
- Read the saved content;
- Send it to an **HTTP server** that should handle it appropriately.

This way the plugin can communicate with any server that has a route to handle this.

For the visual part, we need:

- A **form** where the user can enter the IP and port where the server will be running;
- A **button** to trigger the sync event;
- And, of course, a **debug button**.

# Getting hands dirty

As a starting point, I chose two native KOReader plugins as a base:

- The _Calibre_ one, which would give me a foundation for how to display components on screen;
- And the _Exporter_, which is a plugin already designed to handle highlights.

The first task was to figure out how to start the plugin and show a simple "Hello world" on screen — and for that, the structure below was necessary:

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

Roughly explaining this code:  
we initialized a class called **`ObsidianSync`**, which inherits from the **`WidgetContainer`** class all the basic methods and attributes we need to create a _plugin/widget_ in KOReader.

We also created an **`init`** method, which will be called when the app starts and will be responsible for adding the plugin to the interface.

Finally, we created the **`addToMainMenu`** method which, as the name suggests, is responsible for adding the plugin's visual elements to the menu.

## Event to capture basic information
After creating the basic plugin structure, I needed to figure out how to capture the basic information from the book we're working with, things like title, number of pages, reading progress, and, just as important, the **highlights**.

So, to start this part, I decided to create a **`debug`** method that would return the basic information about which book we're on, what its extension is, and in which directory it's located.

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

There's probably a simpler approach to get the book data, but nothing that a `regex` and some willpower can't solve.

## Accessing the highlights
Now that we have access to the most basic information about the current book, we need to figure out where KoReader saves the information we want from the book.

After a few minutes of research and testing, I ended up seeing that in the same folder where the book is saved, another folder called *filename.sdr* is generated, and in that folder, a file *metadata.{book extension}.lua* is created that contains all the basic information about the book.

Analyzing the metadata file, I found the *annotation* key that contains all the saved **highlights** and their information.

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

## Designing server communication
Now that we have the data we need, we just need to be able to communicate with the server and send the files.

I thought of N ways we could make this connection, but I decided to focus on the basics first and then improve it if necessary. So, to start, we just need the **`IP`** of the machine where the server will be running and the **`PORT`** where the service will be listening.

Before trying to communicate with the server, I decided to create a basic configuration file so the user can fill in this information once and reuse it in future reading sessions.

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

							-- Preserves existing device_id when saving
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

This method is responsible for creating the inputs and saving the data filled in by the user in a configuration file that persists independently of the current book or session.

Now that we have the data for where to send the files, we can just grab the metadata and send it to the server — the metadata is basically an object (table) and we'll let the server handle which data it wants or not.

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

Here we grab all the _metadata_ and, with the user's saved settings, communicate with the server at the **`/sync`** route to complete the _**highlights**_ **sync**.

# Conclusion
We've reached the end of the _plugin_ development (at least the KoReader part) and the complete code is available on my _github_ in the [koreadersync.koplugin](https://github.com/breno5g/koreadersync.koplugin) repository.
```