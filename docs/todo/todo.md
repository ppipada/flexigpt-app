# Project TODO

## Laundry list

- [x] edit message should fall to the editor with attachments/tools so that resend is consistent
- [x] render images and attachments in msg ui
- [x] attachment tools buttons and shortcuts etc dropdowns ui
- [x] slow paste handling
- [x] system prompt delete is via sys dialog. need it to be a modal
- [x] request details are not very useful as of now in the ui. need what was actually sent rather than out version of completion data. may empty out some data things, but others can be given. Most probably need to check response too.

- [x] Details need to pop up as modal
- [x] need info to user on how the url or any attachment for that matter was processed. like was it finally done as text or link or bytes or what.
- [x] If api errors due to token length issues etc, the reply is fully omitted. need to make same as partial response in api abort cases.
- [x] Update ALL modals to use dialog showmodal with backdrop and escape handling.
- [x] url validation check in all modals after novalidate
- [x] Ariakit select for some select items and menus so that we get proper keyboard navigation.
- [x] Tooltips in checkboxes

- [ ] With attachments, token usage display seems more important.
- [ ] Token count in build completion data may be wrong as it doesnt really account for attachments/tool calls etc. Need to rectify the FilterMessagesByTokenCount function.

- [ ] Lots of testing needed for new attachments.

  - [ ] Check debug details.
  - [ ] Attachment follow through across conversations.
  - [ ] cache token usage.
  - [ ] url attachments.

- [ ] Future: May need to decide later that keeping conversation attachment shapshots is better than reattach etc.
- [ ] Future: Check for race between build and fetch for attachment content fetch and how to resolve. For now this is ok to have as the time between these 2 is very small. If we introduce human feedback in between, we ned to look at this closely.
- [ ] Future: Attachments support

  - [ ] doc formats can be in one place
  - [ ] specialized urls like github/gitlab pr's
  - [ ] local git diff/commit history/etc attachments

- attachments

  - [x] attachments separate package.
  - [x] Need image and associated processing in one place
  - [x] any extension that can be read as text in one place
  - [x] web url that is just html fetch
  - [x] url modal needs styling

  - [x] pdf and associated processing in one place. Integrate a pdf text extractor.
  - [x] need to allow dirs too with some good boundaries.
  - [x] for attachments, need to detect mime and type of file etc is backend and ui should use it rather than infering using extension.
  - [x] Skip dot dirs inside the selected dir

  - [x] attachments from prev turns also need to be used. they may need to be freshly hydrated again, but need to be used.
  - [x] input notes should talk about attachments, turns, etc.
  - [ ] get visible text from html needs to be implemented and a new mode raw vs visible text needs to be added.

- [ ] tool enhancements

  - [ ] attached tools to be given back as schema to consumers for exec
  - [ ] need to list openai/anthropic etc given tools so that they can be invoked in api directly
  - [ ] some parallell tools cna be created locally
  - [ ] web search tools better integration needs to be done e2e
