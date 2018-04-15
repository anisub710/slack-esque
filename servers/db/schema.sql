create table if not exists user {
    id int not null auto_increment primary key,
    email varchar(255) not null,
    -- Add indexes to ensure that the email is unique

    -- passhash varbinary(1024) not null,

    userName varchar(255) not null,
    -- Add indexes to ensure that the userName is unique
    firstName varchar(35) null,
    -- check if firstName and lastName can be null
    lastName varchar(35) null,
    photoURL varchar(2083) not null,
    
    --check UNIQUE
    unique(email),
    unique(userName)
}

