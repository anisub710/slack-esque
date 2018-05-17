// "use strict";

const express = require("express");
const mysql = require("mysql");
const app = express();
const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");
const portNum = parseInt(port);

var Channel = require("./models/channel");
var Message = require("./models/message");


if (isNaN(portNum)) {
    throw new Error("port number is not a number");
}

let db = mysql.createPool({
    host: process.env.MYSQL_ADDR,
    database: process.env.MYSQL_DATABASE,
    user: "root",
    password: process.env.MYSQL_ROOT_PASSWORD
});

app.use(express.json());


const SQL_GET_CHANNELS = "select * from channel c " + 
                            "join channel_users cu on c.id = cu.channelid " + 
                            "join users u on u.id = cu.usersid " + 
                            "order by cu.channelid;";
const SQL_INSERT_CHANNEL = "insert into channel (channelname, channeldescription, channelprivate, createdat, creatorid, editedat) "+ 
                        "values (?, ?, ?, ?, ?, ?);";

const SQL_INSERT_MEMBER = "insert into channel_users (channelid, usersid) values " ; 

const SQL_SELECT_CHANNEL = "select * from channel " +
                            "where id = ?;";

const SQL_100_MESSAGES = "select * from users u " +                             
                            "join messages m on m.creatorid = u.id " +
                            "where m.channelid = ? " + 
                            "order by m.createdat " + 
                            "limit 100;";
const SQL_GET_MEMBERS = "select * from channel c " + 
                            "join channel_users cu on cu.channelid = c.id "+
                            "join users u on u.id = cu.usersid " + 
                            "where cu.channelid = ?;";
const SQL_INSERT_MESSAGE = "insert into messages (channelid, body, createdat, creatorid, editedat) "+ 
                            "values (?, ?, ?, ?, ?);";
const SQL_GET_MESSAGE = "select * from messages "+
                            "where id = ?;"
const SQL_UPDATE_CHANNEL = "update channel set channelname = ?, channeldescription = ? where id = ?;";
const SQL_DELETE_MEMBER = "delete from channel_users where channelid = ? and usersid = ?;";
const SQL_UPDATE_MESSAGE = "update messages set body = ? where id = ?;";
const SQL_DELETE_MESSAGE = "delete from messages where id = ?;";
const CONTENT_TYPE = "Content-Type";
const CONTENT_JSON = "application/json";
const CONTENT_TEXT = "text/plain";

function query(db, sql, params) {
    return new Promise((resolve, reject) => {
        db.query(sql, params, (err, results) => {
            if (err) {
                reject(err);
            }else {
                resolve(results);
            }
        });
    });
}


//start the server listening on host:port
app.listen(port, host, () => {
    //callback is executed once server is listening
    console.log(`server is listening at http://${addr}...`);
});

//Handles GET /v1/channels
app.get("/v1/channels", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        let channelsDir = {}; 
        let channels = [];
        db.query(SQL_GET_CHANNELS, (err, rows) => {
            if (err) {
                return next(err);
            }       
 
            rows.forEach((row) => {
                if (!channelsDir[row.channelid]){                                        
                    let newChannel = new Channel(row.channelid, row.channelname, row.channeldescription, row.channelprivate, row.createdat, 
                        {}, row.editedat);  
                    channelsDir[row.channelid] = newChannel;
                }
                if (row.channelprivate){   
                    let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};                      
                    channelsDir[row.channelid].pushMembers(user);
                }
                if (row.creatorid === row.usersid) {
                    let creator = {id: row.creatorid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};
                    channelsDir[row.channelid].setCreator(creator);
                }             

            });
            for (let key in channelsDir){
                let channel = channelsDir[key];                                
                if(channel.private && channel.containsUserID(authResult.id)) {
                    channels.push(channel);                    
                }else if(!channel.private){
                    channels.push(channel);                    
                }
            }                      
            res.setHeader(CONTENT_TYPE, CONTENT_JSON);
            return res.status(200).json(channels);      
        });            
    }
   

});

//Handles POST /v1/channels
app.post("/v1/channels", (req, res, next) => {
    authResult = checkAuthentication(req, res);

    if (req.body.name === "") {
       return res.status(400).send("Provide a name for the channel");
    }

    let time = getTimezoneTime();
    if (authResult) {       
        let newChannel = new Channel(0, req.body.name, req.body.description, req.body.private, time, 
            authResult, time);   
        newChannel.pushMembers(authResult);        
        query(db, SQL_INSERT_CHANNEL, [req.body.name, req.body.description, req.body.private, time, 
            authResult.id, time])
            .then(results => {                
                let newID = results.insertId;  
                newChannel.setId(newID); 
                return newID;               
            })
            .then((newID) => {
                let members = [];
                let memberids = [];
                let params = " (?, ?)";
                memberids.push(newID);
                memberids.push(authResult.id);                 
                if (req.body.private && req.body.members != null){
                    members = req.body.members;
                    for(let i = 0; i < members.length; i++) {
                        params += ",(?, ?)"
                        memberids.push(newID);
                        memberids.push(members[i].id);
                        newChannel.pushMembers(members[i]); 
                    }

                }else if (!req.body.private){ 
                    newChannel.setMembers([]);
                }  
                //used in post for channels/channelID/members                                
                query(db, SQL_INSERT_MEMBER + params, memberids)
                .catch(next);                                                                                                               
                res.setHeader(CONTENT_TYPE, CONTENT_JSON);
                res.status(201).json(newChannel);
            })
            .catch(next);
    }
});

//Handles GET /v1/channels/{channelID}
app.get("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        //TODO:refactor
        db.query(SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if(err) {
               return next(err);
            }
            let result = rows[0];
            let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
            result.createdat, {}, result.editedat);
            //method
            rows.forEach((row) => {
                let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                if(user.id === result.creatorid){
                    channel.setCreator(user);
                }                  
                channel.pushMembers(user);
            });
            
            if (channel.private && !channel.containsUserID(authResult.id)){
                return res.status(403).send("Forbidden access to the channel");
            }
            db.query(SQL_100_MESSAGES, [req.params.channelID], (err, rows) => {
                if (err) {
                    return next(err);
                }                        
                let messages = [];                
                rows.forEach((row) => {
                    let creator = {id: row.creatorid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};                    
                    let message = new Message(row.id, row.channelid, row.body, row.createdat, creator, row.editedat);
                    messages.push(message);
                });
                
                res.setHeader(CONTENT_TYPE, CONTENT_JSON);
                res.status(200).json(messages);          
            });
        });
    }
    
});

//Handles POST /v1/channels/{channelID}
app.post("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){        
        //TODO:refactor
        db.query(SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if(err) {
               return next(err);
            }
            let result = rows[0];
            let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
            result.createdat, {}, result.editedat)
            //method
            rows.forEach((row) => {
                let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                if(user.id === result.creatorid){
                    channel.setCreator(user);
                }                  
                channel.pushMembers(user);
            });
            
            if (channel.private && !channel.containsUserID(authResult.id)){
                return res.status(403).send("Forbidden access to the channel");
            }
            let time = getTimezoneTime();
            //create a new message in this channel 
            let message = new Message(0, req.params.channelID, req.body.body, time, authResult, time);
            db.query(SQL_INSERT_MESSAGE, [message.channelID, message.body, message.createdAt, message.creator.id, message.editedAt], (err, results) => {
                let newID = results.insertId;  
                message.setId(newID);
                res.setHeader(CONTENT_TYPE, CONTENT_JSON);
                res.status(201).json(message);  

            });
        }); 


    }
});

//Handles PATCH /v1/channels/{channelID}
app.patch("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        db.query(SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if (err) {
                return next(err);
            }
            let result = rows[0];               
            let channel = new Channel(req.params.channelID, result.channelname, result.channeldescription, result.channelprivate, 
                result.createdat, {}, result.editedat);

            if (result.creatorid !== authResult.id) {
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }  
            channel.setCreator(authResult);
            if(result.channelprivate){
                rows.forEach((row) => {
                    let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                    channel.pushMembers(user);
                });
            }
            if (req.body.name == null && req.body.description == null) {
                return res.status(400).send("Need to provide either name or description")
            }
            let newName = req.body.name;
            let newDesc = req.body.description;
            if (newName == null && newDesc != null) {
                newName = channel.name;
            }else if (newName != null && newDesc == null){
                newDesc = channel.description;
            }
            db.query(SQL_UPDATE_CHANNEL, [newName, newDesc, req.params.channelID], (err, results) => {
                if (err) {
                    return next(err);
                }
                //check affected rows?
                db.query(SQL_SELECT_CHANNEL, [req.params.channelID], (err, rows) => {
                    if (err) {
                        return next(err);
                    }                    
                    channel.setName(rows[0].channelname);
                    channel.setDescription(rows[0].channeldescription);
                    res.setHeader(CONTENT_TYPE, CONTENT_JSON);
                    res.status(200).json(channel);
                });                
             });
        });
    }    
});

//TODO:Delete
//Handles DELETE /v1/channels/{channelID}
app.delete("/v1/channels/:channelID", (req, res, next) => {
    db.query(SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
        if (err) {
            return next(err);
        }
        let result = rows[0];         
        let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
            result.createdat, {}, result.editedat);

        if (result.creatorid !== authResult.id) {
            return res.status(403).send("Can't make changes to channel since you are not the creator");
        }  
        channel.setCreator(authResult);
        if(result.channelprivate){
            rows.forEach((row) => {
                let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                channel.pushMembers(user);
            });
        }
        
        
    });
});

//TODO: check if private
//TODO: Don't add duplicates   
//Handles POST /v1/channels/{channelID}/members
app.post("/v1/channels/:channelID/members", (req, res, next) => {    
    authResult = checkAuthentication(req, res);
    if(authResult){
        
        //check user is creator
        db.query(SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if (err) {
                return next(err);
            }
            let result = rows[0];         
            let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
                result.createdat, {}, result.editedat);
    
            if (result.creatorid !== authResult.id) {
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }  
            channel.setCreator(authResult);
            if(result.channelprivate){
                rows.forEach((row) => {
                    let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                    channel.pushMembers(user);
                });
            }
            //add user supplied in request body as member of this channel. respond 201 and text user was added
            //only id of user is required. (can post entire profile)
            if (req.body.id == null){
                return res.status(400).send("Should provide a user id to add to channel");
            }
            let addUser = req.body;
            query(db, SQL_INSERT_MEMBER + "(?, ?)", [req.params.channelID, addUser.id])
                .catch(next);                                                                                                               
                res.setHeader(CONTENT_TYPE, CONTENT_TEXT);
                res.status(201).send("Member added to channel " + channel.name);                             
        });
    }
});

//TODO: check if private
//TODO: Don't delete creator?
//TODO: check rows affected?
//Handles DELETE /v1/channels/{channelID}/members
app.delete("/v1/channels/:channelID/members", (req, res, next) => {  
    authResult = checkAuthentication(req, res);
    if(authResult){        
        //check user is creator
        db.query(SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if (err) {
                return next(err);
            }
            let result = rows[0];         
            let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
                result.createdat, {}, result.editedat);
    
            if (result.creatorid !== authResult.id) {
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }  
            channel.setCreator(authResult);
            if(result.channelprivate){
                rows.forEach((row) => {
                    let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                    channel.pushMembers(user);
                });
            }
            //remove user supplied in body from list of channel members
            //200 and text message indicate user removed.
            //only id required
            if (req.body.id == null){
                return res.status(400).send("Should provide a user id to add to channel");
            }
            let addUser = req.body;
            query(db, SQL_DELETE_MEMBER, [req.params.channelID, addUser.id])
            .catch(next);
            res.setHeader(CONTENT_TYPE, CONTENT_TEXT);
            res.status(200).send("Member deleted from channel " + channel.name);
            
        });
    }
});


//Handles PATCH /v1/messages/{messageID}
//change if not in channel anymore?
app.patch("/v1/messages/:messageID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){
        //check if user is creator of message
        db.query(SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
            if(err){
                return next(err);
            }
            let result = rows[0];
            let message = new Message(result.id, result.channelid, result.body, result.createdat, {}, result.editedat);
            if(result.creatorid !== authResult.id) {
                return res.status(403).send("Can't make changes to this message since you are not the creator");
            }
            message.setCreator(authResult);

            if(req.body.body == null){
                return res.status(400).send("No message body provided");
            }
            let newBody = req.body.body;
            db.query(SQL_UPDATE_MESSAGE, [newBody, req.params.messageID], (err, results) => {
                if (err) {
                    return next(err);
                }  
                db.query(SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
                    if (err) {
                        return next(err);
                    }     
                    message.setBody(rows[0].body);
                    res.setHeader(CONTENT_TYPE, CONTENT_JSON);
                    res.status(200).json(message);
                });    
            });                                
        });
        //update body from request (400 if nothing)
        //newly updated message model as response
    }
});

//TODO: check rows affected?
//change if not in channel anymore?
//Handles DELETE /v1/messages/{messageID}
app.delete("/v1/messages/:messageID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){
        //check user is creator of message
        db.query(SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
            if(err){
                return next(err);
            }
            let result = rows[0];
            let message = new Message(result.id, result.channelid, result.body, result.createdat, {}, result.editedat);
            if(result.creatorid !== authResult.id) {
                return res.status(403).send("Can't make changes to this message since you are not the creator");
            }
            message.setCreator(authResult);

            //delete message. plain text message that it was successful            
            db.query(SQL_DELETE_MESSAGE, [req.params.messageID], (err, results) => {
                if (err) {
                    return next(err);
                }  
                res.setHeader(CONTENT_TYPE, CONTENT_TEXT);
                res.status(200).send("Message deleted");
            });
        });
    }
});

function checkUserInChannel(userID, channelID){    

}

app.use((err, req, res, next) => {
    if (err.stack) {
        console.error(err.stack);
    }
    return res.status(500).send("Error in the server");
});

function checkAuthentication(req, res){
    let userJSON = req.get("X-User");
    if (userJSON) {
        let user = JSON.parse(userJSON);        
        return user;
    }else{
        res.status(401).send("Please sign in");
        return null;
    }  
}

function checkCreator(channelID, userID){
    //used in first get. refactor?
    db.query(SQL_SELECT_CHANNEL, [channelID], (err, rows) => {
        //TODO: check error
        if (rows[0].creatorid == userID) {
            return true;
        }        
    });
    return false;
}

//getCurrentTime returns time based on time zone
function getTimezoneTime() {    
    let starttime = new Date();
    let isotime = new Date((new Date(starttime)).toISOString() );
    let fixedtime = new Date(isotime.getTime()-(starttime.getTimezoneOffset()*60000));
    let currentTime = fixedtime.toISOString().slice(0, 19).replace('T', ' ');   
    return currentTime;
}