# FlexiGPT

FlexiGPT is an AI application platform designed to empower organizations in seamlessly developing and deploying AI-driven solutions across their engineering teams. FlexiGPT ensures that your AI initiatives are both scalable and sustainable.

Key objectives of FlexiGPT

- Robust and flexible chat interface
  - Enable dynamic interactions through tools, prompt templates, and knowledge servers, enhancing user engagement and productivity.
  - Support text and code generation APIs from multiple AI providers such as OpenAI, Anthropic, and Google.
- Comprehensive control mechanisms
  - Implement user and cost management controls while maintaining high levels of observability to ensure efficient resource utilization.
- Facilitated development of use-case focused applications
  - Provide a base UI and backend platform for developing use-case focused, LLM powered, applications while maintaining a seamless user experience.
- Cross-Platform Compatibility
  - Ensure seamless operation across Linux, Mac, and Windows systems.

## Install

### Linux

- Download the `.flatpak` release package.
- If flatpak is not enabled, initialize it for the distribution

  - Ubuntu/Debian/etc (APT based systems):

    ```shell
    sudo apt update # update packages
    sudo apt install -y flatpak # install flatpak
    sudo apt install -y gnome-software-plugin-flatpak # optional, enables flathub packages in gnome sofware center
    flatpak remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo
    ```

- Install the package
  - `flatpak install --user FlexiGPT-xyz.flatpak`
  - `flatpak info io.github.flexigpt.client`

### Mac

- Download the `.pkg` release package.
- The current package is not signed using an Apple developer ID based process. Mac would give a security warning for this.
- To install the `.pkg`
  - Click on pkg. It should prompt you to `Move to bin` or `Close`. Choose `Close`.
  - Not Navigate to `Settings` > `Privacy and security`.
  - Scroll to near bottom of the settings and find the `Blocked` section.
  - It should list FlexiGPT there. Click `Open anyway`.
  - It will again ask to confirm. Click `Open` and then install.
