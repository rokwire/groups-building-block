# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased


### Added
- Change participant age field for participant filtering in research groups feature [#604](https://github.com/rokwire/groups-building-block/issues/604)

## [1.71.0] - 2025-08-21
### Fixed
- Not able to add a new staff members NetID to be an admin of a group [#606](https://github.com/rokwire/groups-building-block/issues/606)


## [1.70.1] - 2025-07-21
### Fixed
- Fix nil pointer error [#598](https://github.com/rokwire/groups-building-block/issues/598)

## [1.70.0] - 2025-07-21
### Fixed
- Fix and improve Authman sync logs [#598](https://github.com/rokwire/groups-building-block/issues/598)

## [1.69.0] - 2025-07-16
### Added
- Introduce progressive filter count API [#583](https://github.com/rokwire/groups-building-block/issues/583)

## [1.68.1] - 2025-07-11
### Fixed
- Broken env auth algorithm [#595](https://github.com/rokwire/groups-building-block/issues/595)

## [1.68.0] - 2025-07-10
### Fixed
- Accidental group deletion [#592](https://github.com/rokwire/groups-building-block/issues/564)

## [1.67.0] - 2025-07-08
### Added
- Delete group should delete linked data in other BBs like messages, polls, events as well [#565](https://github.com/rokwire/groups-building-block/issues/564)

## [1.66.1] - 2025-07-10
### Fixed
- Accidental group deletion [#592](https://github.com/rokwire/groups-building-block/issues/564)

## [1.66.0] - 2025-06-24
### Changed
- Update auth library and resolve the active vulnerability [#564](https://github.com/rokwire/groups-building-block/issues/564)


## [1.65.5] - 2025-06-16
### Fixed
- Update groups sort order [#584](https://github.com/rokwire/groups-building-block/issues/584)

## [1.65.4] - 2025-06-16
### Fixed
- Additional minor fix on group's date_updated field. [#566](https://github.com/rokwire/groups-building-block/issues/566)


## [1.65.3] - 2025-06-13
### Fixed
- Additional fix. Don't update group's date_updated field on group settings update. [#566](https://github.com/rokwire/groups-building-block/issues/566)

## [1.65.2] - 2025-06-11
### Changed
- Changed the code owner of the repo.

### Fixed
- Additional fix. Don't update group's date_updated field on a membership update event. [#566](https://github.com/rokwire/groups-building-block/issues/566)

## [1.65.1] - 2025-06-10
### Fixed
- "administrative" field is not handled correctly when creating group [#579](https://github.com/rokwire/groups-building-block/issues/579)

## [1.65.0] - 2025-06-10
### Added
- Add new group type - "administrative" bool [#576](https://github.com/rokwire/groups-building-block/issues/576)

## [1.64.3] - 2025-05-27
### Fixed
- Group Membership - Rejected vs Denied and display discrepancy [#574](https://github.com/rokwire/groups-building-block/issues/574)
- Fix documentation issue [#572](https://github.com/rokwire/groups-building-block/issues/572)


## [1.64.2] - 2025-05-21
### Fixed
- Additional fix on validation [#569](https://github.com/rokwire/groups-building-block/issues/569)

## [1.64.1] - 2025-05-21
### Fixed
- Rename param name to post_update in UpdateGroupDateUpdated api [#569](https://github.com/rokwire/groups-building-block/issues/569)

## [1.64.0] - 2025-05-15
### Added
- Add a description field to the /gr/api/analytics/groups analytics API [#567](https://github.com/rokwire/groups-building-block/issues/567)
- Inactive Groups [#552](https://github.com/rokwire/groups-building-block/issues/552)

## [1.63.0] - 2025-05-09
### Changed
- Improve PUT group/{group-id}/members/multi-create [#562](https://github.com/rokwire/groups-building-block/issues/562)
### Added
- Improve POST api/groups [#561](https://github.com/rokwire/groups-building-block/issues/561)

## [1.62.0] - 2025-04-17
### Changed
- Support Google Trust Services as CA [#558](https://github.com/rokwire/groups-building-block/issues/558)

## [1.61.0] - 2025-04-07
### Changed
- Only the delete managed group operation should require managed_group_admin permission [#553](https://github.com/rokwire/groups-building-block/issues/553)

## [1.60.0] - 2025-03-04
### Changed
- Cleanup and remove the legacy posts logic [#549](https://github.com/rokwire/groups-building-block/issues/549)

## [1.59.1] - 2025-02-12
### Changed
- Additional fix: Load all groups [#536](https://github.com/rokwire/groups-building-block/issues/536)

## [1.59.0] - 2025-02-12
### Changed
- Load all groups [#536](https://github.com/rokwire/groups-building-block/issues/536)


## [1.58.2] - 2025-02-11
### Added
- Customization options to choose group modules: Polls, Direct Messaging, Posts, Events [#545](https://github.com/rokwire/groups-building-block/issues/545)

## [1.58.1] - 2025-02-10
### Fixed
- Members fail to load for some groups on prod [#542](https://github.com/rokwire/groups-building-block/issues/542)


## [1.58.0] - 2025-02-04
### Added
- Client APIs to provide Admin functionality [#541](https://github.com/rokwire/groups-building-block/issues/541)


## [1.57.1] - 2025-02-03
### Changed
- Fix consolidate the information, and make it accessible with a single API call [#538](https://github.com/rokwire/groups-building-block/issues/538)
### Fixed
- Fix migration data permission [#531](https://github.com/rokwire/groups-building-block/issues/531)
- Create posts migration api for switching the datasource to Social BB [#529](https://github.com/rokwire/groups-building-block/issues/529)

## [1.56.0] - 2025-01-23
### Added
- Create migration api for Social BB [#527](https://github.com/rokwire/groups-building-block/issues/527)
- Consolidate the information, and make it accessible with a single API call [#519](https://github.com/rokwire/groups-building-block/issues/519)
- Implement GET /v2/groups API via http POST method [#524](https://github.com/rokwire/groups-building-block/issues/524)

## [1.55.0] - 2024-11-13
### Added 
- BBs API to Get groups by group_ids [#521] (https://github.com/rokwire/groups-building-block/issues/521)

## [1.54.0] - 2024-10-29
### Added 
- Get groups membership by groupID BBs [#516] (https://github.com/rokwire/groups-building-block/issues/516)

## [1.53.0] - 2024-10-22
### Added
- FERPA issues for group memberships [#513](https://github.com/rokwire/groups-building-block/issues/513)

## [1.52.0] - 2024-09-11
### Fixed
- New Event notifications are not sent to Group members [#506](https://github.com/rokwire/groups-building-block/issues/506)

## [1.51.1] - 2024-09-04
### Fixed
- Bad Authman sync for user who has alternative auth method for first login [#509](https://github.com/rokwire/groups-building-block/issues/509)

## [1.51.0] - 2024-08-23
### Added
- Calendar events issues [#503](https://github.com/rokwire/groups-building-block/issues/503)

## [1.50.0] - 2024-08-22
### Added 
- Add "group_id" to the bbs get group membership API [#500](https://github.com/rokwire/groups-building-block/issues/500)

## [1.49.0] - 2024-08-19
### Fixed
- Fix Aggregation pipeline [#497](https://github.com/rokwire/groups-building-block/issues/497)
### Added
- BBs Api for getting group memberships [#494](https://github.com/rokwire/groups-building-block/issues/494)

## [1.48.0] - 2024-08-09
### Added
- Approve all API [#484](https://github.com/rokwire/groups-building-block/issues/484)
### Fixed 
- Fix Groups Stats are not updated [#490](https://github.com/rokwire/groups-building-block/issues/490)

## [1.47.0] - 2024-08-08
### Fixed
- The value for "current_member" is not available in group json [#485](https://github.com/rokwire/groups-building-block/issues/485)
- Groups Stats are not updated [#483](https://github.com/rokwire/groups-building-block/issues/483)
### Added
- POST request for loading group members [#455](https://github.com/rokwire/groups-building-block/issues/455)

## [1.46.3] - 2024-07-24
### Fixed
- Use "get_groups" permission for loading user groups in the admin API [#481](https://github.com/rokwire/groups-building-block/issues/481)

## [1.46.2] - 2024-07-16
### Fixed
- Additional fix: Truncate the post body to 250 characters within the notification[#457](https://github.com/rokwire/groups-building-block/issues/457)

## [1.46.1] - 2024-07-11
### Fixed
- Do not send polls and direct message notifications as muted when they are not [#477](https://github.com/rokwire/groups-building-block/issues/477)

## [1.46.0] - 2024-07-10
### Added
- Admin API for adding group members by NetIDs [#458](https://github.com/rokwire/groups-building-block/issues/458)

## [1.45.2] - 2024-07-01
### Fixed
- Update direct messages notification pattern [#475](https://github.com/rokwire/groups-building-block/issues/475)


## [1.45.1] - 2024-06-26
### Changed
- Use all_bbs_groups & get_aggregated-users permissions for BBs APIs (Additional change) [#473](https://github.com/rokwire/groups-building-block/issues/473)

## [1.45.0] - 2024-06-26
### Changed
- Use all_bbs_groups permissions for all BBs APIs [#473](https://github.com/rokwire/groups-building-block/issues/473)

## [1.44.0] - 2024-06-24
### Added
- Include sender, post content and action in group post notification body [#457](https://github.com/rokwire/groups-building-block/issues/457)
- Send post notification only to the creator of the post [#372](https://github.com/rokwire/groups-building-block/issues/372)

## [1.43.0] - 2024-06-19
### Added
- Provide Replies when loading single Post [#468](https://github.com/rokwire/groups-building-block/issues/468)
- Create Group Report Abuse API [#456](https://github.com/rokwire/groups-building-block/issues/456)

## [1.42.0] - 2024-06-12
### Changed
- Updated golang & alpine Docker container versions
### Added
- Delete everything from the database related to the core account  [#463](https://github.com/rokwire/groups-building-block/issues/463)
- Add new permissions for managing group events independently for granting access [#465](https://github.com/rokwire/groups-building-block/issues/465)

## [1.41.0] - 2024-06-05
### Added
- Introduce BBs APIs. Implement aggregate event users. [#459](https://github.com/rokwire/groups-building-block/issues/459)

## [1.40.2] - 2024-06-03
### Fixed
- Fix missing member name and email for a managed group auto sync task [#460](https://github.com/rokwire/groups-building-block/issues/460)

## [1.40.1] - 2024-05-10
### Fixed
- Additional fix and cleanup[#452](https://github.com/rokwire/groups-building-block/issues/452)

## [1.40.0] - 2024-05-09
### Added
- Improve create group admin api [#452](https://github.com/rokwire/groups-building-block/issues/452)

## [1.39.0] - 2024-04-30
### Added
- Improve group events APIs [#450](https://github.com/rokwire/groups-building-block/issues/450)

## [1.38.0] - 2024-04-26
### Added
- Groups rapid fixes and improvements [#447](https://github.com/rokwire/groups-building-block/issues/447)
- Group Create Post adds members by default [#442](https://github.com/rokwire/groups-building-block/issues/442)
- Implement create and update post admin APIs [#448](https://github.com/rokwire/groups-building-block/issues/448)
- API for creating Group for the Admin app [#445](https://github.com/rokwire/groups-building-block/issues/445)
- API for updating group for the Admin app [#446](https://github.com/rokwire/groups-building-block/issues/446)
- Ability to filter authman/manged groups [#441](https://github.com/rokwire/groups-building-block/issues/441)

## [1.37.1] - 2024-04-24
### Fixed
- Additional fix for scheduled posts[#437](https://github.com/rokwire/groups-building-block/issues/437)

## [1.37.0] - 2024-04-24
### Added
- Schedule post in the future [#437](https://github.com/rokwire/groups-building-block/issues/437)

## [1.36.0] - 2024-04-17
### Added
- Split posts and direct messages within the group [#434](https://github.com/rokwire/groups-building-block/issues/434)

## [1.35.1] - 2024-04-10
### Changed
- Additional fix dead loop & memory leak in the Authman sync task [#428](https://github.com/rokwire/groups-building-block/issues/428)

## [1.35.0] - 2024-04-05
### Changed
- Additional refactor authman automatic sync task[#428](https://github.com/rokwire/groups-building-block/issues/428)

## [1.34.0] - 2024-03-22
### Changed
- Refactor Authman automatic sync task [#428](https://github.com/rokwire/groups-building-block/issues/428)

## [1.33.0] - 2024-03-06
### Added
- Update the schema for Rokwire analytics api for Splunk ingest [#429](https://github.com/rokwire/groups-building-block/issues/429)

## [1.32.1] - 2024-02-28
### Fixed
- Additional fix related to membership & whole group deletion if the user is the only admin [#425](https://github.com/rokwire/groups-building-block/issues/425)

## [1.32.0] - 2024-02-28
### Fixed
- DELETE api/user API does not remove all user activity in groups [#425](https://github.com/rokwire/groups-building-block/issues/425)

## [1.31.0] - 2024-02-15
### Changed
- Group Admin and Event Admin roles should be treated separately [#423](https://github.com/rokwire/groups-building-block/issues/423)

## [1.30.1] - 2024-02-07
### Fixed
- Additional fix for delete event mappings[#417](https://github.com/rokwire/groups-building-block/issues/417)

## [1.30.0] - 2024-02-07
### Added
- Retrieve group ids by event Id [#416](https://github.com/rokwire/groups-building-block/issues/416)
- Update the set of groups that event is published to [#417](https://github.com/rokwire/groups-building-block/issues/417)

## [1.29.0] - 2024-02-06
### Changed
- Do not load "unpublished" events [#414](https://github.com/rokwire/groups-building-block/issues/414)

## [1.28.2] - 2024-02-01
### Fixed
- Additional NPE fix [#411](https://github.com/rokwire/groups-building-block/issues/411)

## [1.28.1] - 2024-02-01
### Added
- Provision all group admins as Event admins [#411](https://github.com/rokwire/groups-building-block/issues/411)

## [1.28.0] - 2024-01-30
### Changed
- Disable automatic event memership provision for group events [#411](https://github.com/rokwire/groups-building-block/issues/411)

## [1.27.3] - 2024-01-05
### Added
- Ability to select event admins when creating event linked to multiple [#408](https://github.com/rokwire/groups-building-block/issues/408)

## [1.27.2] - 2023-11-08
### Changed
- Additional fix of api docs [#405](https://github.com/rokwire/groups-building-block/issues/405)

## [1.27.1] - 2023-11-08
### Changed
- Restructure api doc - client & admin grouping [#405](https://github.com/rokwire/groups-building-block/issues/405)

## [1.27.0] - 2023-11-08
### Added
- New V3 admin API for linking event to set of groups [#403](https://github.com/rokwire/groups-building-block/issues/403)

## [1.26.0] - 2023-10-18
### Changed
- Refactor create and update group logic and resolve the risk of single point of failure. [#401](https://github.com/rokwire/groups-building-block/issues/401)

## [1.25.0] - 2023-10-11
### Fixed
- Fix bad handling of the unique group title index [#399](https://github.com/rokwire/groups-building-block/issues/399)

## [1.24.0] - 2023-10-10
### Fixed
- V3 Load events does not respect time filter [#396](https://github.com/rokwire/groups-building-block/issues/396)

## [1.23.0] - 2023-09-27
### Added
- Implement PUT api/group/{id}/events/v3 API for group calendar events [#394](https://github.com/rokwire/groups-building-block/issues/394)

## [1.22.0] - 2023-09-26
### Added
- Integrate Calendar BB for group events [#392](https://github.com/rokwire/groups-building-block/issues/392)
- Create an adaptor for requesting Calendar BB for dealing with group events [#391](https://github.com/rokwire/groups-building-block/issues/391)

## [1.21.0] - 2023-09-19
### Changed
- Updated libraries and docker container due to vulnerabilities along with the original ticket[#386](https://github.com/rokwire/groups-building-block/issues/386)
### Fixed
- Groups member sorting should be sorting by name [#386](https://github.com/rokwire/groups-building-block/issues/386)

## [1.20.0] - 2023-08-14
### Added
- More Analytics APIs and improvements for getting groups, posts and members [#382](https://github.com/rokwire/groups-building-block/issues/382)

## [1.19.0] - 2023-08-03
### Added
- Analytics API for getting posts [#382](https://github.com/rokwire/groups-building-block/issues/382)
- Prepare for deployment in OpenShift [#379](https://github.com/rokwire/groups-building-block/issues/379)

## [1.18.2] - 2023-05-03
- Fix research groups handling for the internal APIs [#376](https://github.com/rokwire/groups-building-block/issues/376)

## [1.18.1] - 2023-05-02
- Fix bad attributes migration for category & tags [#374](https://github.com/rokwire/groups-building-block/issues/374)

## [1.18.0] - 2023-04-12
### Added
- Improve FCM messages differenciate normal groups and research projects [#370](https://github.com/rokwire/groups-building-block/issues/370)

## [1.17.0] - 2023-04-11
### Changed
- Enable support of can_join_automatically flag for research groups [#368](https://github.com/rokwire/groups-building-block/issues/368)

## [1.16.4] - 2023-04-04
### Fixed
- Fix sending notifications when creating new group [#366](https://github.com/rokwire/groups-building-block/issues/366)

## [1.16.3] - 2023-03-28
### Fixed
- Unable to store group web_url [#364](https://github.com/rokwire/groups-building-block/issues/364)

## [1.16.2] - 2023-03-15
### Fixed
- Internal create event api crashes woth error 500 bug [#360](https://github.com/rokwire/groups-building-block/issues/357)

## [1.16.1] - 2023-02-14
### Fixed
- Fix cast error within Tags backward compatibility handling [#357](https://github.com/rokwire/groups-building-block/issues/357)

## [1.16.0] - 2023-02-07
### Added
- Introduce category and tags backward compatibility [#355](https://github.com/rokwire/groups-building-block/issues/355)

## [1.15.0] - 2023-02-01
### Added
- Add indexes for the nested attributes [#351](https://github.com/rokwire/groups-building-block/issues/351)
### Changed
- Remove category validation on create & update group operations [#352](https://github.com/rokwire/groups-building-block/issues/352)

## [1.14.0] - 2023-01-30
### Added
- Integrate govulncheck within the build process [#319](https://github.com/rokwire/groups-building-block/issues/319)
### Changed
- Rename group filters to attributes [#348](https://github.com/rokwire/groups-building-block/issues/348)

## [1.13.0] - 2023-01-26
### Added
- Implement content filters [#344](https://github.com/rokwire/groups-building-block/issues/344)

## [1.12.4] - 2023-03-20
### Fixed
- Hotfix of [#360] internal API for creating an event (as v1.12.4) [#362](https://github.com/rokwire/groups-building-block/issues/362)

## [1.12.3] - 2023-01-23
### Security
- Use admin token check for delete group membership admin API

## [1.12.2] - 2023-01-23
### Fixed
- Incorrect membership "status" for admins in managed group sync [#341](https://github.com/rokwire/groups-building-block/issues/341)

## [1.12.1] - 2023-01-23
### Fixed
- Additional fixes that improves the admin client APIs [#339](https://github.com/rokwire/groups-building-block/issues/339)

## [1.12.0] - 2023-01-20
### Added
- Add membership retrieve, update & delete admin APIs [#339](https://github.com/rokwire/groups-building-block/issues/339)

## [1.11.0] - 2023-01-13
### Added
- Create internal API for updating group's date updated [#335](https://github.com/rokwire/groups-building-block/issues/335)

## [1.10.1] - 2023-01-10
### Changed
- Split date modified field and introduce member & managed member modified date fields [#330](https://github.com/rokwire/groups-building-block/issues/330)

## [1.9.5] - 2023-01-04
### Changed
- Report of offensive speech automatic title [#328](https://github.com/rokwire/groups-building-block/issues/328)

## [1.9.4] - 2022-12-21
### Changed
- Group admins must not see direct messages if they are not listed explicitly within. the ACL list[#326](https://github.com/rokwire/groups-building-block/issues/326)

## [1.9.3] - 2022-12-20
### Fixed
- Delete group post request fails [#321](https://github.com/rokwire/groups-building-block/issues/321)

## [1.9.2] - 2022-12-14
### Fixed
- Admin must make posts and reactions no matter of group settings [#315](https://github.com/rokwire/groups-building-block/issues/315)

## [1.9.1] - 2022-12-14
### Changed 
- Change the group sorting only to title [#313](https://github.com/rokwire/groups-building-block/issues/313)
### Fixed
- Fix backward compatibility for old clients which don't support group settings [#311](https://github.com/rokwire/groups-building-block/issues/311)


## [1.9.0] - 2022-12-07
### Added
- Add group settings and preferences [#309](https://github.com/rokwire/groups-building-block/issues/309)

## [1.8.0] - 2022-12-01
### Added
- Send notification to potential research group candidates [#298](https://github.com/rokwire/groups-building-block/issues/298)

## [1.7.6] - 2022-11-28
### Fixed
- Fix inappropriate permission check in GET /api/group/{groupId}/posts/{postId} request [#305](https://github.com/rokwire/groups-building-block/issues/305)

## [1.7.5] - 2022-11-23
### Changed
- Upgrade auth library [#303](https://github.com/rokwire/groups-building-block/issues/303)

## [1.7.4] - 2022-11-22
### Changed
- Disable auto join feature for research groups [#301](https://github.com/rokwire/groups-building-block/issues/301)

## [1.7.3] - 2022-11-21
### Fixed
- Count of users matching research profile wrong for empty profile [#299](https://github.com/rokwire/groups-building-block/issues/299)

## [1.7.2] - 2022-11-18
### Added
- Set appID and orgID for notifications [#268](https://github.com/rokwire/groups-building-block/issues/268)

## [1.7.1] - 2022-11-18
### Added
- Add ability to exclude user's groups from response [#295](https://github.com/rokwire/groups-building-block/issues/295)
### Fixed
- Bad title index produce wrong client error for unique title violation [#296](https://github.com/rokwire/groups-building-block/issues/296)

## [1.7.0] - 2022-11-16
### Added
- Add API to get count of users matching research profile [#291](https://github.com/rokwire/groups-building-block/issues/291)
### Changed
- Upgrade auth library [#290](https://github.com/rokwire/groups-building-block/issues/290)

## [1.6.7] - 2022-11-16
### Changed
- Improvement of the original ticket: (research_confirmation is renamed to research_consent_statement and added new field with name research_consent_details)[#288](https://github.com/rokwire/groups-building-block/issues/288)

## [1.6.6] - 2022-11-15
### Added
- Add research_confirmation field within the group [#288](https://github.com/rokwire/groups-building-block/issues/288)
- Bad transaction handling on delete post and delete user data [#287](https://github.com/rokwire/groups-building-block/issues/287)

## [1.6.5] - 2022-11-14
### Fixed
- Wrong property name usage for membership notification preferences for create/update membership records [#285](https://github.com/rokwire/groups-building-block/issues/285)

## [1.6.4] - 2022-11-10
### Fixed
GET /v2/groups doesn't support anonymous users (tokens) [#282](https://github.com/rokwire/groups-building-block/issues/282)

## [1.6.3] - 2022-11-09
### Changed
- Included additional setting to mute all notifications to the original feature [#270](https://github.com/rokwire/groups-building-block/issues/270)
### Fixed
- Fix bad group stats bug [#279](https://github.com/rokwire/groups-building-block/issues/279)

## [1.6.2] - 2022-11-03
### Added
- Override update notifications preferences [#270](https://github.com/rokwire/groups-building-block/issues/270)
- Group notification internal API [#241](https://github.com/rokwire/groups-building-block/issues/241)
- Introduce research groups [#276](https://github.com/rokwire/groups-building-block/issues/276)

## [1.6.1] - 2022-10-12
### Changed
- Finish the transition process of splitting the group record and membership list on two separate collections [#238](https://github.com/rokwire/groups-building-block/issues/238)
- Store group stats [#227](https://github.com/rokwire/groups-building-block/issues/227)

## [1.5.74] - 2022-09-30
### Added
- Allow editing & deleting of a managed groups only if managed_group_admin permission presents for the user [#263](https://github.com/rokwire/groups-building-block/issues/263)

## [1.5.73] - 2022-09-28
### Fixed
- Returning the empty strings for the privacy(public) group members response [#261](https://github.com/rokwire/groups-building-block/issues/261)

## [1.5.72] - 2022-09-27
### Fixed
- Fix Get Group Members [#259](https://github.com/rokwire/groups-building-block/issues/259)

## [1.5.71] - 2022-09-26
### Added
- Set privacy members dependence [#257](https://github.com/rokwire/groups-building-block/issues/257)

## [1.5.70] - 2022-09-19
### Fixed
- Fix inconsistent mix of memberships between the groups & group_membership collections [#252](https://github.com/rokwire/groups-building-block/issues/252)

## [1.5.69] - 2022-09-16
### Added
- Implement additional flag for including hidden groups while searching for group name[#253](https://github.com/rokwire/groups-building-block/issues/253)

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
