alter table `runner` add column `cleaning_request_id` varchar(64) null;
alter table `runner` add column `cleaning_status` varchar(16) null;
alter table `runner` add column `cleaning_error` varchar(512) null;
alter table `runner` add column `cleaning_updated_at` datetime null;
