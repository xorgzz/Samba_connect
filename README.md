# Samba_connect
![image](https://github.com/xorgzz/Samba_connect/assets/118397053/849f7bf5-5e1c-4cea-89e3-ff9d4ea40d58)
Golang TUI app that simplifies mounting samba drives in linux
<br>
### How to set up
`$ sudo apt install libncurses-dev`<br>
`$ git clone https://github.com/xorgzz/Samba_connect`<br>
`$ cd Samba_connect` <br>
`$ go build`<br>
### How to use
There are 2 modes Shell and TUI
#### Shell
Mount:<br>
`$ sudo ./samba_connect user:password local_mount_point:samba_server`<br>
Dismount:<br>
`$ sudo ./samba_connect local_mount_point`
#### TUI
 * q - quit
 * m - mount
 * d - dismount

Data viewd in TUI is pulled from `servers.json`,
user and password can be blank so if they are you'll be prompted for input. 
