from burp import IBurpExtender, ITab
from javax.swing import JPanel, JLabel, JTextField, JButton, JComboBox, JTextArea, JScrollPane, BoxLayout, JOptionPane
from java.awt import BorderLayout, GridBagLayout, GridBagConstraints, Dimension, Color
import threading
import json
import ssl
import urllib2
import base64

try:
    import requests
except ImportError:
    # Burp Jython doesn't have requests, so you may need to bundle or use urllib
    import urllib2 as requests

class BurpExtender(IBurpExtender, ITab):
    def registerExtenderCallbacks(self, callbacks):
        self._callbacks = callbacks
        self._helpers = callbacks.getHelpers()
        callbacks.setExtensionName("Oll0ma File Viewer")

        # Main panel
        self._mainPanel = JPanel(BorderLayout())

        # Top Panel with form
        self._topPanel = JPanel(GridBagLayout())
        self._topPanel.setBackground(Color(220, 224, 233))
        gbc = GridBagConstraints()
        gbc.fill = GridBagConstraints.HORIZONTAL
        gbc.gridy = 0

        self._textFields = []

        labels = ["Queue Server:", "API Key:"]
        for i, label in enumerate(labels):
            gbc.gridx = 0
            gbc.gridy = i
            self._topPanel.add(JLabel(label), gbc)

            gbc.gridx = 1
            textField = JTextField(72)
            self._textFields.append(textField)
            self._topPanel.add(textField, gbc)

        self._textFields[0].setText("https://10.27.20.160:8443")
        self._textFields[1].setText("<Copy from queue server console>")
        
        # Add Submit button
        gbc.gridx = 0
        gbc.gridy = len(labels)
        gbc.gridwidth = 2
        submitButton = JButton("Get Files", actionPerformed=self.handleSubmit)
        submitButton.setBackground(Color(205, 215, 214))
        self._topPanel.add(submitButton, gbc)

        self._mainPanel.add(self._topPanel, BorderLayout.NORTH)
        

        # Dropdown for files
        self._centerPanel = JPanel(BorderLayout())
        
        self.file_dropdown = JComboBox([])
        #self.file_dropdown.add(JLabel("Select File to Pull"), BorderLayout.NORTH)
        self.file_dropdown.addActionListener(self.file_selected)
        self.file_dropdown.setMaximumSize(Dimension(200, 25))
        self._centerPanel.add(self.file_dropdown, BorderLayout.NORTH)

        #self._mainPanel.add(self._filePanel)

        # Text area for file contents
        self.text_area = JTextArea()
        self.text_area.setEditable(False)
        scroll = JScrollPane(self.text_area)
        self._centerPanel.add(scroll, BorderLayout.CENTER)
        self._mainPanel.add(self._centerPanel, BorderLayout.CENTER)

        # Register tab
        callbacks.addSuiteTab(self)

    def getTabCaption(self):
        return "Oll0ma File Viewer"

    def getUiComponent(self):
        return self._mainPanel

    def handleSubmit(self, event):
        def run():
            api_base = self._textFields[0].getText().strip()
            try:
                url = api_base + "/api/files"
                payload = {
                    "apiKey": self._textFields[1].getText().strip()
                }
                jsonData = json.dumps(payload)

                # Send JSON POST
                req = urllib2.Request(url, jsonData)
                req.add_header('Content-Type', 'application/json')

                context = ssl._create_unverified_context()
                resp = urllib2.urlopen(req, context=context)
                JOptionPane.showMessageDialog(
                    self._mainPanel, 
                    "Submitted successfully! HTTP {}".format(resp.getcode())
                )

                #self.text_area.setText(resp.read())
                data = resp.read()
                #self.text_area.append(str(data) + "\n")
                fileJson = json.loads(data.decode('utf-8'))
                #self.text_area.append(str(fileJson) + "\n")

                self.file_dropdown.removeAllItems()
                for f in fileJson['files']:
                    #self.text_area.append(f + "\n")
                    self.file_dropdown.addItem(f)

            except Exception as e:
                JOptionPane.showMessageDialog(self._mainPanel, "Error: {}".format(str(e)))

        threading.Thread(target=run).start()

    def file_selected(self, event):
        if event.getActionCommand() != "comboBoxChanged":
            return
        def run():
            selected_file = self.file_dropdown.getSelectedItem()
            api_base = self._textFields[0].getText().strip()
            try:
                # Change this to a POST
                url = api_base + "/api/file"
                #url = "http://localhost:8443/api/file"
                payload = {
                    "apiKey": self._textFields[1].getText().strip(),
                    "fileName": selected_file
                }
                jsonData = json.dumps(payload)

                req = urllib2.Request(url, jsonData)
                req.add_header('Content-Type', 'application/json')

                context = ssl._create_unverified_context()
                response = urllib2.urlopen(req, context=context)
                JOptionPane.showMessageDialog(
                    self._mainPanel, 
                    "File requested successfully! HTTP {}".format(response.getcode())
                )

                data = response.read()
                fileJson = json.loads(data.decode('utf-8'))
                encryptedData = fileJson['encodedFile']
                decodedData = base64.b64decode(encryptedData).decode('utf-8', errors='ignore')
                
                jsonDecoded = json.loads(decodedData)
                outputModel = "Model: " + jsonDecoded['model'] if 'model' in jsonDecoded else "N/A"
                outputCreated = "Created: " + jsonDecoded['created_at'] if 'created_at' in jsonDecoded else "N/A"
                outputContent = "Result:\n" + jsonDecoded['result_message'] if 'result_message' in jsonDecoded else "N/A"
                outputOriginalSystemPrompt = "Original System Prompt:\n" + jsonDecoded['original_system_prompt'] if 'original_system_prompt' in jsonDecoded else "N/A"
                outputOriginalRequest = "Original Request:\n" + jsonDecoded['original_request'] if 'original_request' in jsonDecoded else "N/A"
                outputOriginalResponse = "Original Response:\n" + jsonDecoded['original_response'] if 'original_response' in jsonDecoded else "N/A"
                #output = outputModel + outputCreated + outputContent
                self.text_area.setText(outputModel)
                self.text_area.append("\n")
                self.text_area.append(outputCreated)
                self.text_area.append("\n\n")
                self.text_area.append(outputContent)
                self.text_area.append("\n\n")
                self.text_area.append(outputOriginalSystemPrompt)
                self.text_area.append("\n\n")
                self.text_area.append(outputOriginalRequest)
                self.text_area.append("\n\n")
                self.text_area.append(outputOriginalResponse)
            except Exception as e:
                self.text_area.setText("Error fetching file:\n" + str(e))

        threading.Thread(target=run).start()
