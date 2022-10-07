# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class GitopsZombies < Formula
  desc ""
  homepage ""
  version "0.0.3"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/raffis/gitops-zombies/releases/download/v0.0.3/gitops-zombies_0.0.3_darwin_arm64.tar.gz"
      sha256 "2a10db7105c3c07f8a53d36848903207fc9f15fcb83fd2c8e1bf34884c588c34"

      def install
        bin.install "gitops-zombies"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/raffis/gitops-zombies/releases/download/v0.0.3/gitops-zombies_0.0.3_darwin_amd64.tar.gz"
      sha256 "49388c5c000cec6414c89633f765c2701f402d0d72d1728779f91f2dcd5dc268"

      def install
        bin.install "gitops-zombies"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/raffis/gitops-zombies/releases/download/v0.0.3/gitops-zombies_0.0.3_linux_arm64.tar.gz"
      sha256 "26465a16a09f7acf7220d306beaa9800b9cef9782f3705525838525585698c07"

      def install
        bin.install "gitops-zombies"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/raffis/gitops-zombies/releases/download/v0.0.3/gitops-zombies_0.0.3_linux_amd64.tar.gz"
      sha256 "a6c839362b4cb0c213b585f6e4cb333a94a99dc04cde298fe110922507361a5a"

      def install
        bin.install "gitops-zombies"
      end
    end
  end

  test do
    system "#{bin}/gitops-zombies -h"
  end
end