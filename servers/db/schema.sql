create table if not exists users (
    id int not null auto_increment primary key,
    email varchar(255) not null,  
    passhash binary(60) not null,
    username varchar(255) not null,
    firstname varchar(35) null,    
    lastname varchar(35) null,
    photourl varchar(2083) not null,
    unique(email),
    unique(username)            
);

