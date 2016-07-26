# csvsql

Golang project to run sql queries on any csv file using embedded sqlite database.  Has interactive and non-interactive modes.

The best way to install is to use the go command:

    `$ go get github.com/niocs/csvsql`

Then add the following line to your .bashrc or .zshrc, etc.

    `export PATH=$GOPATH/bin:$PATH`

```
Usage: csvsql  --Load    <csvfile>
              [--MemDB]             #Creates sqlite db in :memory: instead of disk.
              [--AskType]           #Asks type for each field. Uses TEXT otherwise.
              [--Type  <type1>,...] #Comma separated field types. Can be TEXT/REAL/INTEGER.
              [--TableName]         #Sqlite table name.  Default is csv filename.
              [--Query]             #Query to run. If not provided, enters interactive mode.
              [--RPI]               #Rows per insert. Defaults to 100. Reduce if long rows.
              [--OutFile]           #File to write csv output to. Defaults to stdout.
              [--OutDelim]          #Output Delimiter to use. Defaults is comma.
              [--WorkDir <workdir>] #tmp dir to create db in. Defaults to /tmp/. 
The --WorkDir parameter is ignored if --MemDB is specified.
The --AskType parameter is ignored if --Type  is specified.
The --OutFile parameter is ignored if --Query is NOT specified.
```
