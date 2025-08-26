from burp import IBurpExtender, ITab, IContextMenuFactory
from java.awt import BorderLayout, GridBagLayout, GridBagConstraints
from java.util import ArrayList
from javax.swing import (
    JPanel, JTextArea, JScrollPane, JSplitPane, JMenuItem,
    JLabel, JTextField, JButton, JOptionPane
)
import json
import base64
import urllib2
import ssl

class BurpExtender(IBurpExtender, ITab, IContextMenuFactory):

    def registerExtenderCallbacks(self, callbacks):
        self._callbacks = callbacks
        self._helpers = callbacks.getHelpers()
        callbacks.setExtensionName("Oll0ma Submit to API")

        # === Main Panel ===
        self._mainPanel = JPanel(BorderLayout())

        # === Top Panel with form ===
        self._topPanel = JPanel(GridBagLayout())
        gbc = GridBagConstraints()
        gbc.fill = GridBagConstraints.HORIZONTAL
        gbc.gridy = 0

        self._textFields = []

        labels = ["Queue Server:", "Ollama URL:", "Model:", "API Key:", "RequestNumber:"]
        for i, label in enumerate(labels):
            gbc.gridx = 0
            gbc.gridy = i
            self._topPanel.add(JLabel(label), gbc)

            gbc.gridx = 1
            textField = JTextField(72)
            self._textFields.append(textField)
            self._topPanel.add(textField, gbc)

        self._textFields[0].setText("https://127.0.0.1:8443") # Queue Server URL
        self._textFields[1].setText("") # Ollama URL
        self._textFields[2].setText("") # Model
        self._textFields[3].setText("<Copy from queue server console>") # API Key
        self._textFields[4].setText("0") # Request Number

        # Add Submit button
        gbc.gridx = 0
        gbc.gridy = len(labels)
        gbc.gridwidth = 1
        submitButton = JButton("Load Config", actionPerformed=self.handleLoadConfig)
        self._topPanel.add(submitButton, gbc)

        # Add Load Config Button
        gbc.gridx = 1
        gbc.gridy = len(labels)
        gbc.gridwidth = 1
        loadButton = JButton("Submit", actionPerformed=self.handleSubmit)
        self._topPanel.add(loadButton, gbc)

        self._mainPanel.add(self._topPanel, BorderLayout.NORTH)

        # === Request/Response Panel ===
        self._systemPromptArea = JTextArea()
        systemPromptText = """You are a specialized web application security scanner focused on comprehensive security analysis of web applications and APIs. Your task is to examine the provided information for potential security vulnerabilities, misconfigurations, and architectural weaknesses.

Key areas of focus:
- Architecture & Configuration Review
- Security Headers Analysis
- Authentication & Authorization
- API Security Best Practices
- Common Web Vulnerabilities
- Infrastructure Security

Provide your analysis with clear severity levels using the following format (case-sensitive):
- "**CRITICAL**" for critical security issues
- "**HIGH**" for high-risk vulnerabilities
- "**MEDIUM**" for medium-risk issues
- "**LOW**" for low-risk findings
- "**INFORMATIONAL**" for security observations

Be specific in your findings and include:
1. Clear description of each issue
2. Technical impact
3. Remediation recommendations where applicable

The target information for analysis is provided below this line:"""

        systemPromptText2 = """You are a web application penetration tester conducting a comprehensive operation on an application in the offensive stage of the engagement and focused on leveraging security flaws.
        
        Your objective is to examine the HTTP requests and responses that are available through the burp suite proxy history from the web application as we test the application.
        
        This analysis will focus on:
        - Request and Response Evaluation: Scrutinizing HTTP requests and responses for security misconfigurations, sensitive data exposure, and other vulnerabilities.
        - Authentication and Session Management: Assessing the effectiveness of authentication mechanisms and session handling practices.
        - Input Validation and Output Encoding: Identifying weaknesses related to input validation that may lead to injection attacks or cross-site scripting (XSS).
        
        Use reasoning and context to find potential flaws in the application by providing example payloads and PoCs that could lead to a successful exploit.
        If you deem any vulnerabilities, include the severity of the finding as prepend (case-sensitive) in your response with any of the levels:
        - "**CRITICAL**"
        - "**HIGH**"
        - "**MEDIUM**"
        - "**LOW**"
        - "**INFORMATIONAL**" 
        
        Not every request and response may have any indicators, be concise yet deterministic and creative in your approach.
        The HTTP request and and response pair are provided below this line:
"""
        self._systemPromptArea.setText(systemPromptText2)
        self._requestArea = JTextArea()
        self._responseArea = JTextArea()


        #requestScroll = JScrollPane(self._requestArea)
        #responseScroll = JScrollPane(self._responseArea)

        # Wrap system prompt area with label
        systemPromptPanel = JPanel(BorderLayout())
        systemPromptPanel.add(JLabel("System Prompt"), BorderLayout.NORTH)
        systemPromptPanel.add(JScrollPane(self._systemPromptArea), BorderLayout.CENTER)

        # Wrap request area with label
        requestPanel = JPanel(BorderLayout())
        requestPanel.add(JLabel("Request"), BorderLayout.NORTH)
        requestPanel.add(JScrollPane(self._requestArea), BorderLayout.CENTER)

        # Wrap response area with label
        responsePanel = JPanel(BorderLayout())
        responsePanel.add(JLabel("Response"), BorderLayout.NORTH)
        responsePanel.add(JScrollPane(self._responseArea), BorderLayout.CENTER)

        # First split: request + response
        topSplit = JSplitPane(JSplitPane.VERTICAL_SPLIT, systemPromptPanel, requestPanel)
        topSplit.setDividerLocation(250)

        # Second split: (request+response) + notes
        mainSplit = JSplitPane(JSplitPane.VERTICAL_SPLIT, topSplit, responsePanel)
        mainSplit.setDividerLocation(500)

        # Add everything to main panel
        self._mainPanel.add(mainSplit, BorderLayout.CENTER)

        # Register tab
        callbacks.addSuiteTab(self)

        # Register context menu
        callbacks.registerContextMenuFactory(self)

    #
    # ITab
    #
    def getTabCaption(self):
        return "Oll0ma Submit to API"

    def getUiComponent(self):
        return self._mainPanel

    #
    # IContextMenuFactory
    #
    def createMenuItems(self, invocation):
        menu = ArrayList()
        menu.add(JMenuItem("Oll0ma Submit to API", actionPerformed=lambda x: self.handleSelection(invocation)))
        return menu

    def handleSelection(self, invocation):
        messages = invocation.getSelectedMessages()
        if messages:
            message = messages[0]
            request = message.getRequest()
            response = message.getResponse()

            if request:
                self._requestArea.setText(self._helpers.bytesToString(request))
            else:
                self._requestArea.setText("")

            if response:
                self._responseArea.setText(self._helpers.bytesToString(response))
            else:
                self._responseArea.setText("")

    #
    # Handle the Load Config Button
    # 

    def handleLoadConfig(self, event):
        try:
            url = self._textFields[0].getText().strip()  # Ollama URL from second field
            url = url + "/api/loadconfig"
            if not url:
                JOptionPane.showMessageDialog(self._mainPanel, "Please enter the Ollama URL in Field 2.")
                return

            # Collect the Information for the POST payload
            apiKeyText = self._textFields[3].getText().strip()

            # Construct JSON body
            payload = {
                "apiKey": apiKeyText
            }
            data = json.dumps(payload)
            # Send JSON POST
            req = urllib2.Request(url, data)
            req.add_header("Content-Type", "application/json")

            context = ssl._create_unverified_context()
            resp = urllib2.urlopen(req, context=context)
            data = resp.read()
            self._requestArea.setText(data)
            fileJson = json.loads(data.decode('utf-8'))
            modelString = ""
            for f in fileJson['models']:
                modelString += f + ", "
            modelString = modelString.rstrip(", ")
            modelString = modelString + " - Select 1 from list"
            self._textFields[2].setText(modelString)
            self._textFields[1].setText(fileJson['ollamaURL'])

            JOptionPane.showMessageDialog(self._mainPanel, "Configuration loaded successfully!")

        except Exception as e:
            JOptionPane.showMessageDialog(self._mainPanel, "Error: {}".format(str(e)))        

        

    #
    # Handle Submit Button
    #
    def handleSubmit(self, event):
        try:
            url = self._textFields[0].getText().strip()  # take URL from first field
            url = url + "/api/submit"
            if not url:
                JOptionPane.showMessageDialog(self._mainPanel, "Please enter a URL in Field 1.")
                return

            # Collect text
            requestText = self._requestArea.getText()
            responseText = self._responseArea.getText()
            systemPromptText = self._systemPromptArea.getText()
            requestNumberText = self._textFields[4].getText().strip()
            apiKeyText = self._textFields[3].getText().strip() 
            modelText = self._textFields[2].getText().strip()  

            # Base64 encode
            requestB64 = base64.b64encode(requestText.encode("utf-8"))
            responseB64 = base64.b64encode(responseText.encode("utf-8"))
            systemPromptB64 = base64.b64encode(systemPromptText.encode("utf-8"))
            modelB64 = base64.b64encode(modelText.encode("utf-8"))

            # Construct JSON body
            payload = {
                "apiKey": apiKeyText,
                "model": modelB64,
                "request": requestB64,
                "response": responseB64,
                "systemPrompt": systemPromptB64,
                "requestNumber": requestNumberText
            }
            data = json.dumps(payload)

            # Send JSON POST
            req = urllib2.Request(url, data)
            req.add_header("Content-Type", "application/json")

            context = ssl._create_unverified_context()
            resp = urllib2.urlopen(req, context=context)
            JOptionPane.showMessageDialog(
                self._mainPanel, 
                "Submitted successfully! HTTP {}".format(resp.getcode())
            )

        except Exception as e:
            JOptionPane.showMessageDialog(self._mainPanel, "Error: {}".format(str(e)))
