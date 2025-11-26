# Use git init to initialize it 
# Modify the .git/config file to contain the git SSH URL
# Add your ssh key in the github interface
# git config --global user.email "ltrappett@4gr8.local"

eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519 # Add private key added to the github repo


# To remove an entry in the cache that was added to github
# git rm --cached *.key directory or file that was placed in the .gitignore file

git checkout --orphan temp
git add *
git commit -am "Initial Commit"
git branch -D main
git branch -m main
git push -f origin main
