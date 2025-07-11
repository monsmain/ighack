<div align="center">
<h1>Instructions for importing a personal password file into Termux for tools</h1>
</div>

## Donate❤️
### If you find this code useful, you can support me by donating any amount.
<a href="https://monsmain.github.io/index.html#timeline03-1l"><img src="https://img.shields.io/badge/Donate-E5322D?style=for-the-badge&logo=ilovepdf&logoColor=white" /></a>

## 1. Enable Termux access to the phone's storage
### To allow access, type this command:
```
termux-setup-storage
```
## Then allow access in the window that opens.

## 2. Create a password file:
### Create a text file (for example, named pass.txt) and write the passwords in it, each password on a single line.
### Minimum 6 characters and maximum 64 characters
### Save the file to internal storage, for example (Download folder or any folder).

## 3. Transfer the file to Termux home:
### Run this command to move the file to home:
```
cp /storage/emulated/0/folderx/pass.txt ~/
```
### For example, here we named the folder folderx.
### If the file is somewhere else, enter its path correctly.

## 4. Check that the file has been transferred:

### Run this command and make sure you see pass.txt:
```
ls ~
```
## 5. Then run the code and select the MANUAL ATTACK option and send the file name.
```
../pass.txt
```
### Or
```
pass.txt
```
### After running the commands in the 'Enter Instagram Username' section, write the username and press the enter key.
### And finally, the code runs and finds the user account password.
