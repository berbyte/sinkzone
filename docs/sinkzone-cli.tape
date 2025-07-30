Output examples/demo-cli.gif
Set FontSize 20
Set Height 1000
Set Theme "Rose Pine"
Set Margin 30
Set MarginFill "#6B50FF"
Set Padding 10
Set WindowBar Colorful
Set BorderRadius 10

Type "# Hello, This is sinkzone"
Sleep 4s
Enter

Type "# Let's inspect your current DNS requests"
Sleep 500ms
Enter
Type "sinkzone monitor | tail -10"
Sleep 500ms
Enter
Sleep 4s

Type "# Let's allow github.com"
Sleep 500ms
Enter
Type "sinkzone allowlist add github.com"
Sleep 500ms
Enter
Sleep 4s

Type "# Let's start a focus session"
Sleep 500ms
Enter
Type "sinkzone focus start"
Sleep 500ms
Enter
Sleep 4s

Type "# google.com won't resolve"
Sleep 500ms
Enter
Type "dig google.com"
Sleep 500ms
Enter
Sleep 4s

Type "# github.com is allowed"
Sleep 500ms
Enter
Type "dig github.com"
Sleep 500ms
Enter


Type "Now you can work hard!"
Sleep 10s