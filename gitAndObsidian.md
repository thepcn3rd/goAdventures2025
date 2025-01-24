## Setup git to commit changes
Below are a few of the many references that I have used for this project.  I will also have references inside the code that I publish.

Setup git with SSH - https://www.warp.dev/terminus/git-clone-ssh

After the setup of the above to commit, below is a script I created to automate the commit and remove the history of commits.  The primary reason I remove the history of commits, is in the event I accidentally publish a password, then fix it, someone can not go back in the commit history and pull the password. 

WARNING: This is bad practice to remove the commit history

```bash
# Use git init to initialize it 
# Modify the .git/config file to contain the git SSH URL
# Add your ssh key in the github interface

eval "$(ssh-agent -s)"
# Location of generated key for this project
ssh-add ../sshKeys/goAdventures 

git checkout --orphan temp
git add *
git commit -am "Initial Commit"
git branch -D main
git branch -m main
git push -f origin main
```

#### Setup obsidian to Overlay and Modify Markdown
I utilize obsidian to modify the markdown included in the project.  I create a new vault that is the parent directory of the project and the contents inside the directory.  After doing this in the folder a hidden directory called .obsidian is create.  For good practice I will create a .gitignore file and include the .obsidian file to be ignored

![readme_gitignore.png](/images/readme_gitignore.png)

#### Setup .gitignore file
```text
# Ignore the following directories
.obsidian/
pkg/
bin/
github.com

# Ignore the following file extensions on files
*.bin
*.exe
*.crt
*.key
```
