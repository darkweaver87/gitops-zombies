# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class GitopsZombies < Formula
  desc "Identify kubernetes resources which are not managed by GitOps"
  homepage "https://github.com/raffis/gitops-zombies"
  version "0.0.4"
  depends_on :linux

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/raffis/gitops-zombies/releases/download/v0.0.4/gitops-zombies_0.0.4_linux_arm64.tar.gz"
      sha256 "941170cc9b59b2f411986d10bf2b72ffe2f9b4234a4213591e05ca5041625715"

      def install
        bin.install "gitops-zombies"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/raffis/gitops-zombies/releases/download/v0.0.4/gitops-zombies_0.0.4_linux_amd64.tar.gz"
      sha256 "0681b8dc4c7b19b457390938b61b2e8cada240ab0ca2a2846cd4e2f596b335e2"

      def install
        bin.install "gitops-zombies"
      end
    end
  end

  test do
    system "#{bin}/gitops-zombies -h"
  end
end
