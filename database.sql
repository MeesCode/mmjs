-- running this will empty your old database and create a new one

DROP SCHEMA IF EXISTS Mp3bak2_all;
CREATE SCHEMA IF NOT EXISTS Mp3bak2_all;
USE Mp3bak2_all ;

-- Create table for posts
DROP TABLE IF EXISTS Mp3bak2_all.Folders;

CREATE TABLE IF NOT EXISTS Mp3bak2_all.Folders (
  FolderID int NOT NULL AUTO_INCREMENT,
  Path VARCHAR(255) NOT NULL UNIQUE,
  ParentID int DEFAULT NULL REFERENCES Folders(FolderID),
  PRIMARY KEY (FolderID)
) ENGINE=InnoDB;

-- Create table for threads
DROP TABLE IF EXISTS Mp3bak2_all.Tracks ;

CREATE TABLE IF NOT EXISTS Mp3bak2_all.Tracks (
  TrackID int NOT NULL AUTO_INCREMENT,
  Path VARCHAR(255) NOT NULL UNIQUE,
  FolderID int NOT NULL,
  Title varchar(255) DEFAULT NULL,
	Album varchar(255) DEFAULT NULL,
	Artist varchar(255) DEFAULT NULL,
	Genre varchar(255) DEFAULT NULL,
	Year int DEFAULT NULL,
  PRIMARY KEY (TrackID),
  FOREIGN KEY (FolderID) REFERENCES Folders(FolderID)
) ENGINE=InnoDB;