# csvsql

Golang project to run sql queries on any csv file using embedded sqlite database.  Has interactive and non-interactive modes.

Usage: csvsql  --Load    <csvfile>
              [--MemDB]             #Creates sqlite db in :memory: instead of disk.
              [--AskType]           #Asks type for each field. Uses TEXT otherwise.
              [--TableName]         #Sqlite table name.  Default is t1.
              [--Query]             #Query to run. If not provided, enters interactive mode.
              [--OutFile]           #File to write csv output to. Defaults to stdout.
              [--OutDelim]          #Output Delimiter to use. Defaults is comma.
              [--WorkDir <workdir>] #tmp dir to create db in. Defaults to /tmp/.
The --WorkDir parameter is ignored if --MemDB is specified.
The --OutFile parameter is ignored if --Query is NOT specified.
