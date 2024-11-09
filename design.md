# FlexiGPT: AI App Platform

## Introduction

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

This document outlines the requirements of FlexiGPT, along with a milestone-based task breakdown.

## Requirements

- Chat with LLMs in real-time

  - Support for rich text formatting, including code blocks, mermaid diagrams and mathematical formulas.
  - Edit/Resend previously sent messages.
  - Enable/Disable attaching previous messages to an API call.
  - Tune chat settings such as models/temperature/message control inline.

- Conversation history management

  - Persistent local storage of chats.
  - Ability to view and search past conversations.

- Tools Integration

  - Ability to create and manage custom tools to augment chat functionality.
  - Specify input and output parameters using structured schemas.
  - Create, edit, delete, and duplicate custom tools.
  - Define tools with schemas and executable functions.
  - Support a short code key to access and insert tools in chat

- Prompt Templates

  - Creation and use of predefined prompts within chats.
  - Create, edit, delete, and duplicate prompt templates.
  - Define templates with variables, types, and default values.
  - Support settings config or KB config in prompts.
  - Access and insert templates directly within the chat interface.
  - Option to use templates as system prompts or regular messages.
  - Support a short code key to access and insert prompts in chat

- Knowledge Base Servers

  - Integration with KB servers for in-chat information retrieval.
  - Retrieve and display information within the chat context.
  - Support a short code key to attach a KB in chat
  - Ability to connect to local or remote KB servers

- Use case specific app development

  - A microfrontend based framework to develop usecase specific apps
  - Hooks to integrate the developed apps UI and its backend inside FlexiGPT seamlessly
  - Base utilities and libraries to develop UI as well as backend

- Administration server

  - Ability to connect to a local or remote user control server
  - Mange entities: Org, team, user, secrets
  - Support authentication, authorization of the user
  - Support user preferences, notification
  - Expenditure management and quotas per user/team

- Cross platform applications

  - Develop standalone installers for Linux, Mac, and Windows platforms.
  - Consistent theming across platforms with Dark/Light theme support.

## MileStone 0

- Chat interface

  - Code/mermaid/math/sequence diagram(?) rendering support
  - Edit/Resend support
  - Streaming response support
  - AI Providers integration: OpenAI, Anthropic, Google

- Conversations

  - Local conversation history JSON store via API
  - Supporting search and load

- Settings

  - Local settings JSON store via API
  - Tune settings inline of chat

- Dark/Light theme support
- Local installer for Linux

## MileStone 1

- Tools

  - CRUD + search in a dedicated page
  - In chat usage
  - Provide an editor for defining input/output schemas.
  - Ensure the editor supports schema syntax highlighting and validation.
  - Local tool JSON store via API
  - Short code key support

- Prompts

  - CRUD + search in a dedicated page.
  - Provide an editor for defining prompt template.
  - Support variables and settings config in templates.
  - Associate tools or KB servers with the prompt if needed.
  - In chat usage as system or user prompt, supporting variable population and assignment
  - Ensure the editor supports schema syntax highlighting and validation.
  - Local prompt JSON store via API
  - Short code key support

- Knowledge base servers integration

  - Provide a user interface to add, query and display KB servers.
  - Integrate with KB server APIs.
  - Allow users to retrieve information from KB servers within chats.
  - Short code key support
  - Allow upload/delete docs
  - Create searchable entities within a doc set

- Local installer for Mac/Win (unsigned)

## Milestone 2

- Current feature enhancements

  - CI: Build and distribute packages like DMG (Mac), MSI (Windows), and Flatpak (Linux).
  - Chat interface:
    - Develop better algorithms for deducing conversation titles.
  - Tools: Incorporate code interpretation capabilities for select programming languages as tools.
  - KB stores: Enhance interaction with KB stores beyond basic retrieval.
  - Support for Additional AI Providers
    - Amazon Bedrock
    - Mistral AI
    - Ollama
    - IBM Watsonx
    - LamaFile
    - Cohere
    - Cloudflare
    - Baidu Ernie
    - X Groq

- APM
  - Support token consumption observability
  - Support streamed response as well as direct response tracking
  - Inline indication of token usage and cost probability in chat interface
  - Integrate debugging information to assist in application maintenance.

## Milestone 3

- Usecase specific app support

  - Introduce support for usecase specific apps with dedicated micro UIs and backend.
  - Provide LLM interaction specific UI components
  - Extensible hooks for integration with the platform UI (E.g: MicroUIs/Module federation in NextJS)
  - Include common UI components, state management tools, and API integration utilities.
  - Ways to integrate UI theme in the micro frontend
  - Offer a set of base libraries and utilities to streamline the development of use-case specific apps.
  - Explore presenting App outputs in formats like Jupyter notebooks.

## Milestone 4

- Organizational Admin Server
  - Entity manager: Org, Team, User
  - Authn/Authz support
  - Secrets and Preferences manager
  - Notifications support
  - Expenditure manager

## Possible Sample Apps to Integrate

- Sample Note app

  - A simple markdown based note taking app developed on the platform
  - Support a continuous buffer of block of information.
  - Quick LLM text generate call
  - Calculator mode
  - Use LLMs to organize buffers into locally hosted MD files per task.

- Testing Assistant App

  - Design interfaces for API test creation and management.
  - Implement backend services to support testing workflows.
