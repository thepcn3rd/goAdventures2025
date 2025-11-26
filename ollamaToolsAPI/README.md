# ollama and openwebui

At the SANS DFIR Summit keynote in Salt Lake City, Mari DeGrazia presented _"DFIR AI-ze Your Workflow,"_ introducing the use of **Ollama** and **Open WebUI** to locally run and manage AI models. She demonstrated how to leverage tools for enriching data—such as adding context to an IP address—to streamline investigative workflows. Inspired by her talk, I documented my own journey: from setting up Ollama and Open WebUI on my local machine to migrating it to a Proxmox server in my home lab. These notes cover configuring a **"Model Preset"** and testing its functionality with an external tool emulating an API. Whether for research, coding, or task automation, this walk through aims to help others deploy their own **private LLM** efficiently. (This is a rough draft... Thank you!)

## Setup ollama and openwebui in Docker
On my local laptop I already had docker installed on Ubuntu linux.  After a quick google search I found a docker compose script that worked for what I needed.  The script will create 2 dockers respectively for ollama and openwebui.  I created a directory called ollama and placed this script inside:

```docker
version: '3.8'

services:
  ollama:
    container_name: ollama
    image: ollama/ollama:latest
    ports:
      - "11434:11434" # Expose Ollama API port
    volumes:
      - ollama_data:/root/.ollama # Persistent storage for models
    restart: unless-stopped

  open-webui:
    container_name: open-webui
    image: ghcr.io/open-webui/open-webui:main
    ports:
      - "3000:8080" # Expose Open WebUI web interface
    environment:
      - OLLAMA_BASE_URL=http://ollama:11434 # Connect to Ollama service
    volumes:
      - openwebui_data:/app/backend/data # Persistent storage for Open WebUI data
    depends_on:
      - ollama # Ensure Ollama starts first
    restart: unless-stopped

volumes:
  ollama_data:
  openwebui_data:

```

Notice the port for the openwebui is modified from port 8080 to 3000 and the configuration for it to connect to ollama is in the environment setting.  Below is the command to execute the docker compose file.

```bash
docker compose up -d
```

NOTE: The openwebui runs on the protocol of HTTP but it is not using TLS.


## Configure Open WebUI and Download Models
After connecting to http://127.0.0.1:3000 and clicking at the bottom "Get Started", you are presented with a screen similar to the below.  Your first time will ask you to setup the admin user.

![createAdminAccount](/ollama/picts/createAdminAccount.png)

After the initial login it will look like the below. 

![signinOpenWebUI.png](/ollama/picts/signinOpenWebUI.png)

After authentication you will see in the top-left the ability to add a model to use however no models exist at the moment.

![selectAModel.png](/ollama/picts/selectAModel.png)

To load a model select the top-left menu options and then the left-side panel will appear with the option of the admin panel.  Then select the admin panel.

![adminPanel.png](/ollama/picts/adminPanel.png)

Below is what the admin panel looks like, select settings, then in the left select models.  In the far right is a download icon.  This is where you add (download) or remove models that you are using.

![downloadModels.png](/ollama/picts/downloadModels.png)

Then you will notice a "click here" that takes you to the ollama library of models.  By default it is sorted by the most popular.  You can click on the models and understand them more and search for them.  My 3 interests in models at the moment are for learning, automation with AI and AI assisted coding.

![manageModels.png](/ollama/picts/manageModels.png)

After finding the name of the model that I would like to download copy the name into "Enter model tag" then click the download button next to the tag.  You need to wait until a green box appears in the top-right and the model tag disappears from the box for it to be complete.

I pulled the following models for testing:
1. deepseek-r1
2. deepseek-coder-v2
3. WhiteRabbitNeo (Heard about this from Kindo and was interesting in the security angle)
4. qwen2.5-coder

After loading the models back on the main screen in the top-left you can select the model to use and you can have 2 run in parallel.  This worked well for the deepseek ones but then I loaded the WhiteRabbitNeo and it errored out that I needed more memory.

The usage of the models were slow and I was loosing patience waiting on the results.  I also need to work with and understand the models better to match what my expectations are and my needs.

## Proxmox and OpenWebUI

After loosing some patience and running out of memory on my local computer, I powered up a proxmox server and decided to google if I could load openwebui as a native container.  The following github site for Proxmox VE Helper-Scripts came across the screen.  Due to being paranoid I looked at the script first and then ran the command on my proxmox.

![proxmoxHelperScript.png](/ollama/picts/proxmoxHelperScript.png)

Note that the helper script as you follow the execution, you need to tell it "Y" to install ollama.  Ollama and Open WebUI are installed in the same container.  Also, the docker compose script above changes the port 8080 to 3000, this leaves the port at 8080.

The container will automatically start after the script executes with 4 vCPU, 8GB of RAM and 25GB of disk space.  I then went back and modified the openwebui.sh script manually and increased the disk space to 125GB and re-ran it to create a container.  I increased the disk space due to some of the models average about 6GB in size and due to experimentation I wanted to be able to load 10 or more.

Then I could have modified the vCPUs and the Memory.  I found on my proxmox the most memory that 2 models needed in parallel was about 20GB as I observed the resources.  Because this proxmox did not have a GPU to utilize it was heavy on CPUs.  I provided it 16 vCPUs and at times it was using upwards of 60%.  These settings could have been included in the openwebui.sh file but I manually modified them as I learned how it was going to perform on the proxmox that I have.

## Workspaces and Customizing Models

One of my initial prompts for deepseek coder was to build code to base64 decode a string.  This is a relatively simple task just to test it.  However one of my curiosities sparked by the keynote was to be able to use tools.   

![verifyingCodeBuilding.png](/ollama/picts/verifyingCodeBuilding.png)

Before learning how to use tools I needed to understand and configure workspaces.  I will let the documentation of Open WebUI explain workspaces while I skim over the details that I had to learn.  

First, I clicked on "Discover a model".  I was initially confused because I had already downloaded 3 models above.  These are "Model Presets".

![modelPreset.png](/ollama/picts/modelPreset.png)

After reaching the openwebui.com models location I typed in "code" and observed a Codewriter.  Then I clicked on that "model".

![selectModel.png](/ollama/picts/selectModel.png)

After clicking and scrolling down I observed the "system prompt" that describes how this model present is going to behave and what it thinks it does best as a system prompt.  It does show which Base Model ID it is from so you could replicate and use the same if performance and expectations are met.  I loved seeing at the bottom, "never explain the code just write code".

![modelPresetCodeBuilder.png](/ollama/picts/modelPresetCodeBuilder.png)

Then clicking "GET" I was prompted the URL of my Open WebUI and then by clicking import to WebUI it was completed.  You are prompted 

![importModelPreset.png](/ollama/picts/importModelPreset.png)

Because the Model of llama3 was not found I need to select the model that I am going to use, that I previously downloaded.

![selectBaseModel.png](/ollama/picts/selectBaseModel.png)

Then as I scrolled down I observed where I can select "tools" to use with this "model preset".  At the moment I am not worried about tools.  I deselected "Vision" and "Image Generation" due to not needing them at the moment. (Note: After exploring a couple of models, I gave up on image generation...)  Then click on "Save and Create"

![creatingModelPreset.png](/ollama/picts/creatingModelPreset.png)

Now I have a "model" under workspaces that I can select by clicking on it or modify the settings or the system prompt by editing it.

![modifyModelPresetCodewriter.png](/ollama/picts/modifyModelPresetCodewriter.png)

I clicked on "codewriter" and then created a prompt to generate a function as shown in the image below.

![testingModelPreset.png](/ollama/picts/testingModelPreset.png)
This was an unexpected outcome of exploring this as "codewriter" created the function without explaining it and did not insert comments into the blocks of code to explain it.  Wahoo!!  Anyways back to the goal of using tools.

## Creating a "Model Preset to Run Tools"

After working with tools a little I will shortcut about 3-4 hours of explorations that I did.  First I need to create a "Model Preset" that has a system prompt of the following.

```txt
You are a cyber security analyst that is focused on gathering information about IP Addresses.  Only output the information provided by the tools in a table.  The table has the tool and then the output from the tool.  
```

This is a very simplified version to allow the output from a prompt to focus on using the tools attached to the "Model Preset".  I also downloaded a new "model" called "wizard2llm:7b" and attached to this, again just finding the one that meets my needs and is a little faster (Go and explore...).

![createIPAnalystModel.png](/ollama/picts/createIPAnalystModel.png)


## Creating the Initial Tool
I became lost in the interface a couple of times so I placed the below screenshot here.  Click on Workspace, then "Tools".  I clicked on discover and found a couple of easy tools to study and learn from and then I crafted my own called Get_F1.

![createTool.png](/ollama/picts/createTool.png)

The following python code I included for the tool, also note that a good description helps.  Note: You can have more than 1 function.  Note that all the tool does is, if an ip address is used in a prompt for evaluation then it will call this URL and output either "Malicious", "Safe" or "Bad Request" based on the web server that I build in the next section to emulate an API Endpoint.

```python
import os
import requests
from datetime import datetime
from pydantic import BaseModel, Field

apiURL = "http://10.27.20.197:8000"

class Tools:
	def __init__(self):
		pass

	def f1(self, ipaddr: str) -> str:
		"""
		Analyze an IP Address and Find VirusTotal Information
		:param ipaddr: The IP Address to evaluate
		"""

		url = apiURL + "/f1?ipaddress=" + ipaddr
		response = requests.get(url)
		return getResponse(url)

	def f2(self, ipaddr: str) -> str:
		"""
		Analyze an IP Address and Find Geo Information
		:param ipaddr: The IP Address to evaluate
		"""

		url = apiURL + "/f2?ipaddress=" + ipaddr
		response = requests.get(url)
		return getResponse(url)

	def f3(self, ipaddr: str) -> str:
		"""
		Analyze an IP Address and Find Related Hostnames
		:param ipaddr: The IP Address to evaluate
		"""

		url = apiURL + "/f3?ipaddress=" + ipaddr
		return getResponse(url)

	def getResponse(url: str) -> str:
		response = requests.get(url)
		if response.status_code == 200:
			return response.text
		else:
			return "Bad Request"
```


## Emulating an API End-point

I wrote with the assistance of AI, a web server that would serve as an endpoint emulating a return value of if an IP Address is requested by visiting http://x.x.x.x:8080/f1?ipaddress=5.5.5.5 a response of "Malicious" or "Safe" would be returned.  Again my goal was only to have the "tool" call this URL emulating an API end-point.

![golangCodeEmulateAPI.png](/ollama/picts/golangCodeEmulateAPI.png)

The code above is an example.  Located [here](/ollama/toolCode.txt) is code that will do a lookup with VirusTotal, a GeoLocation query of ip-addr and a reverse dns lookup on the IP Address.  

## Configure the Model Preset to Use the Tool
Click on workspaces and then the edit on the IP Analyst Model.  Then scroll down and select the tool that we created.  Save and Update

![selectTool.png](/ollama/picts/)
![selectTool.png](/ollama/picts/selectTool.png)

Then I started the golang webserver to emulate an API endpoint and it has verbose output of an IP Address and the User-Agent to indicate a connection.

## Create Prompt

Typed, "Is 9.5.4.3 safe?" with the IP Analyst selected.  Waiting for it to think and then query and provide the output.

![promptOutputSafe.png](/ollama/picts/promptOutputSafe.png)

The output shows that the tool executed and returned that the IP Address is safe based on the tool.  Let's double-check our emulation to see the connection.

![connectionToAPI.png](/ollama/picts/connectionToAPI.png)

Based on the output I can see the prompt through the model used the tool.  The request came and it returned that the IP Address is safe.  Let's run it again where the IP Address is 5.5.5.5.  The tool output should show that it is malicious. (Not that it is malicious it was an easy address to test with.)

![promptOutputMal.png](/ollama/picts/)
![promptOutputMal.png](/ollama/picts/promptOutputMal.png)

It worked! The model used the tool!!!  

Through my experience I have 3 functions in 1 tool and I have tested with 1 function in 3 tools.  I found that 3 functions in 1 tool works the best.  I liked the output of deepseek:r1latest the most.  I tried working with other models to speed the processing on my local computer but even though they were faster to execute the tool they were not patient enough to process the data.

Enjoy! Use responsibly!

[License](/LICENSE.md) - Remember based on the agreement that you entered into that SANS book information should not be placed in AI.


