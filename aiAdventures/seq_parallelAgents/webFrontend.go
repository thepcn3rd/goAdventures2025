package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

func applyTemplate(i InformationStruct) string {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <title>Create Email</title>
  <style>
    /* Table with single-line separators between rows and columns */
    table.form-table {
      width: 100%;
      border-collapse: collapse;         /* ensures single borders */
      table-layout: fixed;               /* keeps columns consistent */
    }
    table.form-table th,
    table.form-table td {
      border: 1px solid #ccc;            /* single line separators */
      padding: 8px;
      vertical-align: middle;
    }
    table.form-table th {
      text-align: right;                  /* label column alignment */
      width: 220px;                       /* adjust label column width as needed */
      background: #f7f7f7;
      font-weight: 600;
    }
    table.form-table input[type="text"],
    table.form-table input[type="password"] {
      width: 640px;
      box-sizing: border-box;
      padding: 6px 8px;
    }
    .actions {
      text-align: right;
      padding-top: 12px;
    }
  </style>
</head>
<body>

<form enctype="multipart/form-data" action="/create" method="post" aria-label="Create Email">
  <table class="form-table">
    <tbody>
      <tr>
        <th><label for="txtApikey">API Key:</label></th>
        <td><input type="password" id="txtApikey" name="txtApikey" value="hack" width="350" /></td>
      </tr>
      <tr>
        <th><label for="txtDemographic">Demographic:</label></th>
        <td><input type="text" id="txtDemographic" name="txtDemographic" value="%s" /></td>
      </tr>
      <tr>
        <th><label for="txtURL1">URL #1:</label></th>
        <td><input type="text" id="txtURL1" name="txtURL1" value="%s" /></td>
      </tr>
	  <tr>
        <th><label for="txtURLImage1">URL Image #1:</label></th>
        <td><input type="text" id="txtURLImage1" name="txtURLImage1" value="%s" /></td>
      </tr>
      <tr>
        <th><label for="txtURL2">URL #2:</label></th>
        <td><input type="text" id="txtURL2" name="txtURL2" value="%s" /></td>
      </tr>
	  <tr>
        <th><label for="txtURLImage2">URL Image #2:</label></th>
        <td><input type="text" id="txtURLImage2" name="txtURLImage2" value="%s" /></td>
      </tr>
      <tr>
        <th><label for="txtURL3">URL #3:</label></th>
        <td><input type="text" id="txtURL3" name="txtURL3" value="%s" /></td>
      </tr>
	  <tr>
        <th><label for="txtURLImage3">URL Image #3:</label></th>
        <td><input type="text" id="txtURLImage3" name="txtURLImage3" value="%s" /></td>
      </tr>
      <tr>
        <th><label for="txtURL4">URL #4:</label></th>
        <td><input type="text" id="txtURL4" name="txtURL4" value="%s" /></td>
      </tr>
      <tr>
        <th><label for="txtURLImage4">URL Image #4:</label></th>
        <td><input type="text" id="txtURLImage4" name="txtURLImage4" value="%s" /></td>
      </tr>
      <tr>
        <th><label for="txtCategory">Category URL:</label></th>
        <td><input type="text" id="txtCategory" name="txtCategory" value="%s" /></td>
      </tr>
      <tr>
        <th class="actions" colspan="2" style="text-align: center; border-left: 1px solid #ccc; border-right: 1px solid #ccc;">
          <input type="submit" value="Create Email" />
        </th>
      </tr>
    </tbody>
  </table>
</form>
`, i.DemographicInfo, i.DemographicInfo, i.URL1, i.URLImage1, i.URL2, i.URLImage2, i.URL3, i.URLImage3, i.URL4, i.URLImage4, i.CategoryURL)

	return html

}

func ProvideInformationHTML(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprint(w, headerHTML())
	infoStruct := InformationStruct{
		DemographicInfo: "a husband searching for a christmas gift",
		URL1:            "",
		URLImage1:       "",
		URL2:            "",
		URLImage2:       "",
		URL3:            "",
		URLImage3:       "",
		URL4:            "",
		URLImage4:       "",
		CategoryURL:     "",
	}

	ufHTML := applyTemplate(infoStruct)
	fmt.Fprint(w, ufHTML)
	fmt.Fprint(w, tailHTML())
}

func CreateIndexHTML(folderDir string) {
	currentDir, _ := os.Getwd()
	newDir := currentDir + folderDir
	//cf.CheckError("Unable to get the working directory", err, true)
	if _, err := os.Stat(newDir); errors.Is(err, os.ErrNotExist) {
		// Output to File - Overwrites if file exists...
		f, err := os.Create(newDir)
		if err != nil {
			log.Printf("Unable create file index.html "+currentDir, err)
		}
		defer f.Close()
		f.Write([]byte(headerHTML()))
		f.Write([]byte("AI Learning 2025<br /><br />"))
		f.Write([]byte("<a href='/input.html'>Link to Create Email - College Student</a><br /><br />"))
		f.Write([]byte("<a href='/input2.html'>Link to Create Email - Grandmother of College Students</a><br /><br />"))
		f.Write([]byte("<a href='/input3.html'>Link to Create Email - Husband searching for Christmas Gifts</a><br /><br />"))
		f.Write([]byte(tailHTML()))
		f.Close()
	}
}

func headerHTML() string {
	hHTML := fmt.Sprintln(`<!DOCTYPE html>
<html lang="en">
	<head>
    	<meta charset="UTF-8" />
    	<meta name="viewport" content="width=device-width, initial-scale=1.0" />
    	<meta http-equiv="X-UA-Compatible" content="ie=edge" />
  	</head>
  	<body>`)
	return hHTML
}

func tailHTML() string {
	tHTML := fmt.Sprintln("</body></html>")
	return tHTML
}
