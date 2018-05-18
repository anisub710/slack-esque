module.exports = {
    SQL_GET_CHANNELS: "select * from channel c " + 
                        "join channel_users cu on c.id = cu.channelid " + 
                        "join users u on u.id = cu.usersid " + 
                        "order by cu.channelid;",
    SQL_INSERT_CHANNEL: "insert into channel (channelname, channeldescription, channelprivate, createdat, creatorid, editedat) "+ 
                        "values (?, ?, ?, ?, ?, ?);",
    SQL_INSERT_MEMBER: "insert into channel_users (channelid, usersid) values ", 
    SQL_SELECT_CHANNEL: "select * from channel where id = ?",
    SQL_100_MESSAGES: "select * from users u " +                             
                        "join messages m on m.creatorid = u.id " +
                        "where m.channelid = ? " + 
                        "order by m.createdat " + 
                        "limit 100;",
    SQL_GET_MEMBERS: "select * from channel_users cu " + 
                        "join users u on u.id = cu.usersid " + 
                        "join channel c on c.id = cu.channelid "+                        
                        "where c.id = ?;",
    SQL_INSERT_MESSAGE: "insert into messages (channelid, body, createdat, creatorid, editedat) "+ 
                            "values (?, ?, ?, ?, ?);",
    SQL_GET_MESSAGE: "select * from messages "+
                            "where id = ?;",
    
    SQL_UPDATE_CHANNEL: "update channel set channelname = ?, channeldescription = ? where id = ?;",
    SQL_DELETE_MEMBER: "delete from channel_users where channelid = ? and usersid = ?;",
    SQL_UPDATE_MESSAGE: "update messages set body = ? where id = ?;",
    SQL_DELETE_MESSAGE: "delete from messages where id = ?;",
    SQL_DELETE_MESSAGE: "delete from messages where id = ?;",
    CONTENT_TYPE: "Content-Type",
    CONTENT_JSON: "application/json",
    CONTENT_TEXT: "text/plain"                
}














