from burp import IBurpExtender, ITab, IContextMenuFactory
from java.awt import BorderLayout, GridBagLayout, GridBagConstraints, Color, Dimension
from java.util import ArrayList
from javax.swing import (
    JPanel, JTextArea, JScrollPane, JSplitPane, JMenuItem,
    JLabel, JTextField, JButton, JOptionPane, JComboBox
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

        labels = ["Queue Server:", "API Key:", "Ollama URL:"]
        for i, label in enumerate(labels):
            gbc.gridx = 0
            gbc.gridy = i
            self._topPanel.add(JLabel(label), gbc)

            gbc.gridx = 1
            textField = JTextField(72)
            self._textFields.append(textField)
            self._topPanel.add(textField, gbc)

        self._textFields[0].setText("https://10.27.20.160:8443") # Queue Server URL
        self._textFields[1].setText("<Copy from queue server console>") # API Key
        self._textFields[2].setText("") # Ollama URL

        # Add dropdown list of models to select from
        # Model Label
        gbc.gridx = 0
        gbc.gridy = len(labels)
        gbc.gridwidth = 1
        self._topPanel.add(JLabel("Model:"), gbc)

        # Model Dropdown
        gbc.gridx = 1
        self.model_dropdown = JComboBox([])
        self.model_dropdown.setMaximumSize(Dimension(200,25))
        self._topPanel.add(self.model_dropdown, gbc)

        # Add Load Config button
        gbc.gridx = 0
        gbc.gridy = len(labels)+1
        gbc.gridwidth = 1
        loadButton = JButton("Load Config", actionPerformed=self.handleLoadConfig)
        loadButton.setBackground(Color(184, 216, 216))
        self._topPanel.add(loadButton, gbc)

        # Add Submit Button
        gbc.gridx = 1
        gbc.gridy = len(labels)+1
        gbc.gridwidth = 1
        submitButton = JButton("Submit", actionPerformed=self.handleSubmit)
        submitButton.setBackground(Color(205, 215, 214))
        #submitButton.setMaximumSize(Dimension(50, 30))
        self._topPanel.add(submitButton, gbc)

        self._mainPanel.add(self._topPanel, BorderLayout.NORTH)

        # === Request/Response Panel ===
        self._systemPromptArea = JTextArea()
        systemPromptText = "Loads from configuration file on the oll0ma Queue Server..."
        self._systemPromptArea.setText(systemPromptText)
        self._requestArea = JTextArea()
        self._responseArea = JTextArea()


        #requestScroll = JScrollPane(self._requestArea)
        #responseScroll = JScrollPane(self._responseArea)

        # Wrap system prompt area with label
        systemPromptPanel = JPanel(BorderLayout())
        systemPromptPanel.setBackground(Color(131, 201, 244))
        systemPromptLabel = JLabel("System Prompt")
        systemPromptLabel.setHorizontalAlignment(JLabel.CENTER)
        systemPromptPanel.add((systemPromptLabel), BorderLayout.NORTH)
        systemPromptPanel.add(JScrollPane(self._systemPromptArea), BorderLayout.CENTER)

        # Wrap request area with label
        #requestPanel = JPanel(BorderLayout())
        #requestPanel.add(JLabel("Request"), BorderLayout.NORTH)
        #requestPanel.add(JScrollPane(self._requestArea), BorderLayout.CENTER)

        requestTopPanel = JPanel(GridBagLayout())
        requestTopPanel.setBackground(Color(163, 213, 255))
        c = GridBagConstraints()
        c.gridx = 0; c.gridy = 0; c.gridwidth = 3; c.weightx = 1; c.fill = c.HORIZONTAL
        request_label = JLabel("Request")
        request_label.setHorizontalAlignment(JLabel.CENTER)
        requestTopPanel.add(request_label, c)

        c.gridy = 1; c.gridx = 0; c.gridwidth = 1; c.weightx = 0.30
        middle_label = JLabel("Request Number:")
        middle_label.setHorizontalAlignment(JLabel.RIGHT)
        requestTopPanel.add(middle_label, c)
        #requestTopPanel.add(JLabel("Request Number:"), c)

        c.gridx = 1; c.weightx = 0.30
        self._requestNumberField = JTextField("0",10)
        requestTopPanel.add(self._requestNumberField, c)

        c.gridx = 2; c.weightx = 0.4
        sidePanel = JPanel()
        sidePanel.setBackground(Color(163, 213, 255))
        requestTopPanel.add((sidePanel), c)

        requestPanel = JPanel(BorderLayout())
        #requestPanel.add(JLabel("Request"), BorderLayout.NORTH)
        requestPanel.add(requestTopPanel, BorderLayout.NORTH)
        requestPanel.add(JScrollPane(self._requestArea), BorderLayout.CENTER)
        

        # Wrap response area with label
        responsePanel = JPanel(BorderLayout())
        responsePanel.setBackground(Color(217, 240, 255))
        responseLabel = JLabel("Response")
        responseLabel.setHorizontalAlignment(JLabel.CENTER)
        responsePanel.add((responseLabel), BorderLayout.NORTH)
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
            apiKeyText = self._textFields[1].getText().strip()

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
            #self._requestArea.setText(data)
            fileJson = json.loads(data.decode('utf-8'))
            
            # Populate the Model dropdown
            self.model_dropdown.removeAllItems()
            for m in fileJson['models']:
                self.model_dropdown.addItem(m)
            
            self._textFields[2].setText(fileJson['ollamaURL'])
            self._systemPromptArea.setText(fileJson['systemPrompt'])

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
            requestNumberText = self._requestNumberField.getText().strip()
            apiKeyText = self._textFields[1].getText().strip() 
            modelText = self.model_dropdown.getSelectedItem()  

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
