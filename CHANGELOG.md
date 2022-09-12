# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

## [1.5.68] - 2022-09-12
### Fixed
- Exception on get group posts [#250](https://github.com/rokwire/groups-building-block/issues/250)

## [1.5.67] - 2022-09-12
### Added
- Implement an API that returns a single Post [#244](https://github.com/rokwire/groups-building-block/issues/244)
- Add reactions to posts [#242](https://github.com/rokwire/groups-building-block/issues/242)
- Expose DELETE /api/int/group/{group-id}/events/{event-id} API for using by Events Manager [#236](https://github.com/rokwire/groups-building-block/issues/236)
- Retrieve members by list of account IDs [#246](https://github.com/rokwire/groups-building-block/issues/246)

## [1.5.66] - 2022-08-24
### Fixed
- Fix wrong groups pagination [#232](https://github.com/rokwire/groups-building-block/issues/232)

## [1.5.65] - 2022-08-17
### Fixed
- Maximum Mongo document size limit for groups [#222](https://github.com/rokwire/groups-building-block/issues/222)

## [1.5.64] - 2022-08-10
### Fixed
- Authman groups members missing [#218](https://github.com/rokwire/groups-building-block/issues/218)

## [1.5.63] - 2022-08-09
### Fixed
- Fix merge issues

## [1.5.62] - 2022-08-08
### Added
- Improve managed group admin assignment and synchronization [#209](https://github.com/rokwire/groups-building-block/issues/209)

## [1.5.61] - 2022-08-05
### Added
- Introduce V2 group APIs and improve the legacy with additional filter options  [#212](https://github.com/rokwire/groups-building-block/issues/212)
  - Introduced V2 APIs (the members list is omitted from the v1 response):
    - GET /api/v2/groups
    - GET /api/v2/groups/{id}
    - GET /api/v2/user/groups
    - GET /api/admin/v2/groups
    - GET /api/admin/v2/groups/{id}
    - GET /api/admin/v2/user/groups
    - GET /api/admin/group/{group-id}/stats
    - GET /api/group/{id}/stats
    - GET /api/group/{group-id}/members

## [1.5.60] - 2022-08-03 
- Test build

## [1.5.59] - 2022-07-29
### Added
- Create internal API for creating a group event by another BB [#210](https://github.com/rokwire/groups-building-block/issues/210)

## [1.5.58] - 2022-07-27
### Added
- Add internal API for retrieving group members by group title [#205](https://github.com/rokwire/groups-building-block/issues/205)

## [1.5.57] - 2022-07-26
### Changed
- Improve report a group post as an abuse [#204](https://github.com/rokwire/groups-building-block/issues/204)

## [1.5.56] - 2022-07-22
### Added
- Introduce admin Authman sync api (POST /admin/authman/synchronize) [#202](https://github.com/rokwire/groups-building-block/issues/202)
### Fixed
- Improve logging of the internal API calls [#200](https://github.com/rokwire/groups-building-block/issues/200)

## [1.5.55] - 2022-07-21
### Fixed
- Authman sync task should add default admins only to the new groups [#198](https://github.com/rokwire/groups-building-block/issues/198)

## [1.5.54] - 2022-07-19
### Changed
- Check the group is eligible for autumn synchronisation before initiate the operation [#196](https://github.com/rokwire/groups-building-block/issues/196)

## [1.5.53] - 2022-07-18
### Changed
- Change the default title of groups transactional FCM messages to "Group - {Group Name}" [#194](https://github.com/rokwire/groups-building-block/issues/194)

## [1.5.52] - 2022-07-15
### Changed
- Set "Academic" category for all Gies auto created groups [#192](https://github.com/rokwire/groups-building-block/issues/192)
### Fixed
- Fix admin authorization [#190](https://github.com/rokwire/groups-building-block/issues/190)

## [1.5.51] - 2022-07-13
### Fixed
- [BUG-UIUC] Groups - editing group settings kicks out all members [#188](https://github.com/rokwire/groups-building-block/issues/188)

## [1.5.50] - 2022-07-12
### Changed
- Expose netID within the membership & user records [#184](https://github.com/rokwire/groups-building-block/issues/163)
- Deprecate ROKWIRE_GS_API_KEY and start using only INTERNAL-API-KEY as internal API authentication mechanism. This  a redo operations of a previous redo changes due to a confirmation [#156](https://github.com/rokwire/groups-building-block/issues/156)

## [1.5.49] - 2022-07-07
### Fixed
- Additional fix for missing member.id on requesting for a group membership [#163](https://github.com/rokwire/groups-building-block/issues/163)

## [1.5.48] - 2022-07-06
### Changed
- Prepare the project to become open source [#146](https://github.com/rokwire/groups-building-block/issues/146)
### Fixed
- Additional fix for missing client_id on creating a new group [#163](https://github.com/rokwire/groups-building-block/issues/163)
- Additional fix for missing member creation date on requesting for a group membership [#163](https://github.com/rokwire/groups-building-block/issues/163)

## [1.5.47] - 2022-07-01
### Changed
- Handle Autumn group pretty name and the default admins [#177](https://github.com/rokwire/groups-building-block/issues/177)

## [1.5.46] - 2022-06-30
### Changed
- Improve report abuse email template [#174](https://github.com/rokwire/groups-building-block/issues/174)

## [1.5.45] - 2022-06-24
### Changed
- Internal Autumn synch API needs to take parameters for stem checks [#167](https://github.com/rokwire/groups-building-block/issues/167)

## [1.5.44] - 2022-06-21
### Fixed
- Additional fix for missing ID on new member request [#163](https://github.com/rokwire/groups-building-block/issues/163)
- Clean polls logic and remove it from the Groups BB [#150](https://github.com/rokwire/groups-building-block/issues/150)

## [1.5.43] - 2022-06-16
### Added
- Add the group member to the autumn group automatically  [#163](https://github.com/rokwire/groups-building-block/issues/163)

## [1.5.42] - 2022-06-08
### Added
- Report a post as abuse [#161](https://github.com/rokwire/groups-building-block/issues/161)

## [1.5.41] - 2022-06-06
### Added
- Implement ability to use post's subject and body as a notification for the all group members [#159](https://github.com/rokwire/groups-building-block/issues/159)

## [1.5.40] - 2022-06-02
### Changed
- Rollback previous changes and add support of the both ROKWIRE_GS_API_KEY & INTERNAL-API-KEY headers for backward compatibility [#156](https://github.com/rokwire/groups-building-block/issues/156)

## [1.5.39] - 2022-06-01
### Changed
- Deprecate ROKWIRE_GS_API_KEY and start using INTERNAL-API-KEY as internal API authentication mechanism [#156](https://github.com/rokwire/groups-building-block/issues/156)

## [1.5.38] - 2022-05-30
### Added
- Add default Authman group admins on Authman group creation[#153](https://github.com/rokwire/groups-building-block/issues/153)

## [1.5.37] - 2022-05-27
### Added
- Implement automatic Authman group creation and membership synchronisation [#153](https://github.com/rokwire/groups-building-block/issues/153)

## [1.5.36] - 2022-05-26
### Added
- Create internal group detail API [#151](https://github.com/rokwire/groups-building-block/issues/151)

## [1.5.35] - 2022-05-18
### Added
- Add support for attendance groups [#147](https://github.com/rokwire/groups-building-block/issues/147)

## [1.5.34] - 2022-05-16
### Added
- Ability to initiate manual Authman synch by group admin [#144](https://github.com/rokwire/groups-building-block/issues/143)

## [1.5.33] - 2022-05-11
### Added
- Add support of closed groups [#143](https://github.com/rokwire/groups-building-block/issues/143)

## [1.5.32] - 2022-05-05
### Fixed
- Additional fix due to broken UIUC client [#132](https://github.com/rokwire/groups-building-block/issues/132)

## [1.5.31] - 2022-05-04
### Fixed 
- Additional fixes for subgroup notifications for posts, polls and events [#140](https://github.com/rokwire/groups-building-block/issues/140)

### Added
- Migrate group polls mapping and move ot to the Groups BB side [#140](https://github.com/rokwire/groups-building-block/issues/140)

## [1.5.30] - 2022-04-28
### Fixed
- Fix auth library usage issues [#138](https://github.com/rokwire/groups-building-block/issues/138)

## [1.5.29] - 2022-04-26
### Changed
- Update Swagger library due to security issue [#135](https://github.com/rokwire/groups-building-block/issues/135)

## [1.5.28] - 2022-04-20
### Added
- Hide groups from search queries with additional flag [#132](https://github.com/rokwire/groups-building-block/issues/132)

## [1.5.27] - 2022-04-14
### Fixed
- Anonymous users need access /groups API too [#130](https://github.com/rokwire/groups-building-block/issues/130)

## [1.5.26] - 2022-04-04
### Added
- Implement additional GET /group/{group-id}/events/v2 api to avoid breaking existing clients [#128](https://github.com/rokwire/groups-building-block/issues/126)

## [1.5.25] - 2022-03-23
### Added
- Add support of members_to for events [#128](https://github.com/rokwire/groups-building-block/issues/126)

## [1.5.24] - 2022-03-10
### Fixed
- Authman settings are not serialised if the group is private and the user is not a member [#126](https://github.com/rokwire/groups-building-block/issues/126)

## [1.5.24] - 2022-03-10
### Fixed
- Authman settings are not serialised if the group is private and the user is not a member [#126](https://github.com/rokwire/groups-building-block/issues/126)

## [1.5.23] - 2022-03-09
### Fixed
- More fixes and improvements for post destination members restriction [#123](https://github.com/rokwire/groups-building-block/issues/123)

## [1.5.22] - 2022-03-08
### Changed
- Limit reply to a specific subset of members [#123](https://github.com/rokwire/groups-building-block/issues/123)

## [1.5.21] - 2022-02-24
### Added
- Additional fix of wrong membership initialization [#122](https://github.com/rokwire/groups-building-block/issues/122)

## [1.5.20] - 2022-02-23
### Changed
- Deprecate and remove member.user.id and use member.user_id [#122](https://github.com/rokwire/groups-building-block/issues/122)

## [1.5.19] - 2022-02-16
### Changed
- Ignore members different from uofinetid [#116](https://github.com/rokwire/groups-building-block/issues/116)

## [1.5.18] - 2022-02-15
### Added
- Additional improvements for Authman sync [#116](https://github.com/rokwire/groups-building-block/issues/116)

## [1.5.17] - 2022-02-14
### Added
- Add additional parameter within the authman user api [#116](https://github.com/rokwire/groups-building-block/issues/116)

## [1.5.16] - 2022-02-11
### Changed
- Add only_admins_can_create_posts flag to the group object [#116](https://github.com/rokwire/groups-building-block/issues/116)

## [1.5.15] - 2022-02-09
### Fixed
- Additional fixes for Login & Authman sync [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.14] - 2022-02-09
### Fixed
- Additional fixes for Login & Authman sync [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.13] - 2022-02-08
### Fixed
- Additional fixes for Login & Authman sync [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.12] - 2022-02-07
### Added
- Additional fixes for Login & Authman sync [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.11] - 2022-02-03
### Fixed
-Additional fixes for Login & Authman sync [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.10] - 2022-02-02
### Fixed
- Add internal stats API [#112](https://github.com/rokwire/groups-building-block/issues/112)

## [1.5.9] - 2022-01-31
### Fixed
- Authman sync fixes and improvements [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.8] - 2022-01-27
### Fixed
- Authman sync fixes and improvements [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.7] - 2022-01-12
### Fixed
- Authman sync fixes and improvements [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.6] - 2022-01-10
- Authman sync fixes and improvements [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.6] - 2022-01-10
### Changed
- Update core auth library and cache the name of the user for Authman sync pupronse [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.5.5] - 2022-01-05 
### Changed
- Updated changelog

## [1.5.4] - 2022-01-04 (Rejected)
### Added
- Implement CRUD APIs for Poll mappings [#107](https://github.com/rokwire/groups-building-block/issues/107)

## [1.5.3] - 2021-12-27
### Added
- Add Authman API support [#106](https://github.com/rokwire/groups-building-block/issues/106)

## [1.4.44] - 2021-12-17
### Fixed
- Fix wrong check for admin while getting posts for group [#103](https://github.com/rokwire/groups-building-block/issues/103)

## [1.4.43] - 2021-12-15
### Fixed
- Fix inconsistent is_core_user flag for core user[#100](https://github.com/rokwire/groups-building-block/issues/100)

## [1.4.42] - 2021-12-08
### Added
- Add image_url in the Post model[#96](https://github.com/rokwire/groups-building-block/issues/96)

## [1.4.41] - 2021-12-10
## [1.4.40] - 2021-12-09

## [1.4.39] - 2021-12-08
### Fixed
Fail with error 401 or 403 during the authentication & authorization phase [#91](https://github.com/rokwire/groups-building-block/issues/91)

## [1.4.38] - 2021-12-06
### Fixed
- Don't send notifications to all users [#89](https://github.com/rokwire/groups-building-block/issues/89)
