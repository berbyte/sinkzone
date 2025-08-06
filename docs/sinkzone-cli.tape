Output examples/demo-cli.gif
Set FontSize 20
Set Height 1000
Set Theme "Rose Pine"
Set Margin 30
Set MarginFill "#6B50FF"
Set Padding 10
Set WindowBar Colorful
Set BorderRadius 10

Hide
Type "sinkzone allowlist remove github.com && clear"
Enter
Show


Type "# Hello, this is sinkzone! 👋"
Enter
Type "# Let's inspect your current DNS requests"
Sleep 500ms
Enter

Type "sinkzone monitor | head -10"
Sleep 500ms
Enter
Sleep 2s

Type "# Let's allow github.com"
Sleep 500ms
Enter
Type "sinkzone allowlist add github.com"
Sleep 500ms
Enter
Sleep 2s

Type "# Let's start a focus session"
Sleep 500ms
Enter
Type "sinkzone focus start"
Sleep 500ms
Enter
Sleep 2s

Type "# google.com won't resolve"
Sleep 500ms
Enter
Type "dig google.com"
Sleep 500ms
Enter
Sleep 2s

Type "# github.com is allowed"
Sleep 500ms
Enter
Type "dig github.com"
Sleep 500ms
Enter


Type "Enjoy the silence! :)"
Sleep 10s
