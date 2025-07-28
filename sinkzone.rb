class Sinkzone < Formula
  desc "A strict DNS filter to help you stay focused — or keep your kids safe"
  homepage "https://github.com/berbyte/sinkzone"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/berbyte/sinkzone/releases/download/v#{version}/sinkzone_Darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_ARM64" # This will be updated by GoReleaser
    else
      url "https://github.com/berbyte/sinkzone/releases/download/v#{version}/sinkzone_Darwin_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_X86_64" # This will be updated by GoReleaser
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/berbyte/sinkzone/releases/download/v#{version}/sinkzone_Linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_ARM64" # This will be updated by GoReleaser
    else
      url "https://github.com/berbyte/sinkzone/releases/download/v#{version}/sinkzone_Linux_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_X86_64" # This will be updated by GoReleaser
    end
  end

  def install
    bin.install "sinkzone"
  end

  def post_install
    # Create configuration directory
    config_dir = "#{ENV["HOME"]}/.sinkzone"
    unless Dir.exist?(config_dir)
      system "mkdir", "-p", config_dir
      system "chmod", "755", config_dir
    end

    # Create default configuration if it doesn't exist
    config_file = "#{config_dir}/sinkzone.yaml"
    unless File.exist?(config_file)
      File.write(config_file, <<~EOS)
        mode: normal
        upstream_nameservers:
          - "8.8.8.8"
          - "1.1.1.1"
      EOS
      system "chmod", "644", config_file
    end

    # Create state file if it doesn't exist
    state_file = "#{config_dir}/state.json"
    unless File.exist?(state_file)
      File.write(state_file, <<~EOS)
        {
          "focus_mode": false,
          "last_updated": "#{Time.now.iso8601}"
        }
      EOS
      system "chmod", "644", state_file
    end

    # Install systemd service on Linux
    if OS.linux?
      service_dir = "/etc/systemd/system"
      service_file = "#{service_dir}/sinkzone-resolver.service"
      
      unless File.exist?(service_file)
        File.write(service_file, <<~EOS)
          [Unit]
          Description=Sinkzone DNS Resolver
          After=network.target
          Wants=network.target

          [Service]
          Type=simple
          User=root
          ExecStart=#{bin}/sinkzone resolver
          Restart=always
          RestartSec=5
          StandardOutput=journal
          StandardError=journal

          [Install]
          WantedBy=multi-user.target
        EOS
        system "chmod", "644", service_file
        puts "Systemd service installed at #{service_file}"
        puts "To enable and start the service:"
        puts "  sudo systemctl enable sinkzone-resolver"
        puts "  sudo systemctl start sinkzone-resolver"
      end
    end

    # Install launchd service on macOS
    if OS.mac?
      plist_dir = "/Library/LaunchDaemons"
      plist_file = "#{plist_dir}/run.ber.sinkzone.resolver.plist"
      
      unless File.exist?(plist_file)
        File.write(plist_file, <<~EOS)
          <?xml version="1.0" encoding="UTF-8"?>
          <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
          <plist version="1.0">
          <dict>
              <key>Label</key>
              <string>run.ber.sinkzone.resolver</string>
              <key>ProgramArguments</key>
              <array>
                  <string>#{bin}/sinkzone</string>
                  <string>resolver</string>
              </array>
              <key>RunAtLoad</key>
              <true/>
              <key>KeepAlive</key>
              <true/>
              <key>StandardOutPath</key>
              <string>/var/log/sinkzone-resolver.log</string>
              <key>StandardErrorPath</key>
              <string>/var/log/sinkzone-resolver.log</string>
          </dict>
          </plist>
        EOS
        system "chmod", "644", plist_file
        puts "Launchd service installed at #{plist_file}"
        puts "To enable and start the service:"
        puts "  sudo launchctl load #{plist_file}"
      end
    end
  end

  test do
    system "#{bin}/sinkzone", "status"
  end

  def caveats
    caveats_text = <<~EOS
      Sinkzone has been installed successfully!

      To get started:

      1. Start the DNS resolver (requires root):
         sudo sinkzone resolver

      2. Open the TUI in another terminal:
         sinkzone

      3. Enable focus mode (1 hour):
         Press 'f' in the TUI or run:
         sinkzone focus 1h

      4. Check status:
         sinkzone status

      Configuration is stored in ~/.sinkzone/
    EOS

    if OS.linux?
      caveats_text += <<~EOS

        Service Management (Linux):
        - Enable service: sudo systemctl enable sinkzone-resolver
        - Start service: sudo systemctl start sinkzone-resolver
        - Stop service: sudo systemctl stop sinkzone-resolver
        - Check status: sudo systemctl status sinkzone-resolver
      EOS
    elsif OS.mac?
      caveats_text += <<~EOS

        Service Management (macOS):
        - Enable service: sudo launchctl load /Library/LaunchDaemons/run.ber.sinkzone.resolver.plist
        - Stop service: sudo launchctl unload /Library/LaunchDaemons/run.ber.sinkzone.resolver.plist
        - Check logs: tail -f /var/log/sinkzone-resolver.log
      EOS
    end

    caveats_text += <<~EOS

      For more information, visit: https://github.com/berbyte/sinkzone
    EOS

    caveats_text
  end
end 