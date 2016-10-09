-- +migrate Up
CREATE TABLE recordings (
  id UUID,
  file_name TEXT,
  CONSTRAINT pk_recordings PRIMARY KEY (id)
);


-- +migrate Down
DROP TABLE recordings;
