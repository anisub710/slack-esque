// "use strict";

const express = require("express");
const mysql = require("mysql");
const app = express();
const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");
const portNum = parseInt(port);

var Channel = require("./models/channel");
var Message = require("./models/message");
var Constants = require("./constants");

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
        db.query(Constants.SQL_GET_CHANNELS, (err, rows) => {
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

            channels = populateChannels(channelsDir);
            res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
            return res.status(200).json(channels);      
        });            
    }
   

});

function populateChannels(channelsDir) {
    let channels = [];
    for (let key in channelsDir){
        let channel = channelsDir[key];                                
        if(channel.private && channel.containsUserID(authResult.id)) {
            channels.push(channel);                    
        }else if(!channel.private){
            channels.push(channel);                    
        }
    } 
    return channels;
}

//Handles POST /v1/channels
app.post("/v1/channels", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {   
        if (req.body.name === "") {
            return res.status(400).send("Provide a name for the channel");
            }
                
        let newChannel = cleanRequest(req);
        newChannel.pushMembers(authResult);     

        query(db, Constants.SQL_INSERT_CHANNEL, [newChannel.name, newChannel.description, newChannel.private, newChannel.createdAt, 
            authResult.id, newChannel.editedAt])
            .then(results => {                
                let newID = results.insertId;  
                newChannel.setId(newID); 
                return newID;               
            })
            .then((newID) => {                
                let membersMap = buildParams(req, newID, authResult, newChannel);
                newChannel = membersMap.newChannel;                              
                query(db, Constants.SQL_INSERT_MEMBER + membersMap.params, membersMap.ids)
                .catch(next);                                                                                                               
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                res.status(201).json(newChannel);
            })
            .catch(next);
    }
});

function cleanRequest(req){
    let description = "Template description";
    if (req.body.description !== undefined){
        description = req.body.description;
    }

    let setPrivate = false;
    if (req.body.private !== undefined) {
        setPrivate = req.body.private;
    }
    let time = getTimezoneTime();
    let newChannel = new Channel(0, req.body.name, description, setPrivate, time, 
        authResult, time); 
    return newChannel;
}

function buildParams(req, newID, authResult, newChannel){
    let membersMap = {};
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
    membersMap.params = params;
    membersMap.ids = memberids;
    membersMap.newChannel = newChannel;
    return membersMap;
}

//Handles GET /v1/channels/{channelID}
app.get("/v1/channels/:channelID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        try{            
            let inChannel = await checkUserInChannel(req, authResult);
            if (!inChannel) {
                return res.status(403).send("Forbidden access to the channel");
            }
            db.query(Constants.SQL_100_MESSAGES, [req.params.channelID], (err, rows) => {
                if (err) {
                    return next(err);
                }                        
                let messages = [];                
                rows.forEach((row) => {
                    let creator = {id: row.creatorid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};                    
                    let message = new Message(row.id, row.channelid, row.body, row.createdat, creator, row.editedat);
                    messages.push(message);
                });
                
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                res.status(200).json(messages);          
            });  
        }catch(err) {
            next(err);
        }                                  
    }
    
});



//Handles POST /v1/channels/{channelID}
app.post("/v1/channels/:channelID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){                
        //checks if channel has user
        try{                     
            let inChannel = await checkUserInChannel(req, authResult);
            if (!inChannel) {
                return res.status(403).send("Forbidden access to the channel");
            }

            let time = getTimezoneTime();
            //create a new message in this channel 
            let message = new Message(0, req.params.channelID, req.body.body, time, authResult, time);
            db.query(Constants.SQL_INSERT_MESSAGE, [message.channelID, message.body, message.createdAt, message.creator.id, message.editedAt], (err, results) => {
                let newID = results.insertId;  
                message.setId(newID);
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                res.status(201).json(message);  

            });    
        }catch(err) {
            next(err);
        }

    }
});

function checkUserInChannel(req, authResult){
    return new Promise((resolve, reject) => {
        db.query(Constants.SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if (err) {
                reject(err);
            }else {
                let result = rows[0];
                let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
                result.createdat, {}, result.editedat);            
                rows.forEach((row) => {
                    let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                    if(user.id === result.creatorid){
                        channel.setCreator(user);
                    }                  
                    channel.pushMembers(user);
                });
            
                if (channel.private && !channel.containsUserID(authResult.id)){
                    resolve(false);
                }
                resolve(true);
            }
        });
    });
}

//Handles PATCH /v1/channels/{channelID}
app.patch("/v1/channels/:channelID", async (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        
        try{
            //checks if user is creator of channel
            let creator = await checkIsCreator(req, authResult);
            if (!creator) {
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }            
            let channel = await getChannelMembers(req, authResult);
            
            let newName = req.body.name;
            let newDesc = req.body.description;
            if (newName == undefined && newDesc != undefined) {
                newName = channel.name;
            }else if (newName != undefined && newDesc == undefined){
                newDesc = channel.description;
            }
        
            db.query(Constants.SQL_UPDATE_CHANNEL, [newName, newDesc, req.params.channelID], (err, results) => {
                if (err) {
                    return next(err);
                }
                //check affected rows?
                db.query(Constants.SQL_SELECT_CHANNEL, [req.params.channelID], (err, rows) => {
                    if (err) {
                        return next(err);
                    }                    
                    channel.setName(rows[0].channelname);
                    channel.setDescription(rows[0].channeldescription);
                    res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
                    res.status(200).json(channel);
                });                
            });               
        }catch(err) {
            next(err);
        }
                        
    }    
});

function checkIsCreator(req, authResult) {
    return new Promise((resolve, reject) => {
        db.query(Constants.SQL_SELECT_CHANNEL, [req.params.channelID], (err, rows) => {
            if(err) {
                reject(err);
            }else {
                let result = rows[0];
                if (result.creatorid !== authResult.id) {
                    resolve(false);
                }  
                resolve(true);
            }
        });
    });
}

function getChannelMembers(req, authResult) {
    return new Promise((resolve, reject) => {
        db.query(Constants.SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
            if (err) {
                reject(err);
            }else {
                let result = rows[0];         
                let channel = new Channel(result.id, result.channelname, result.channeldescription, result.channelprivate, 
                    result.createdat, {}, result.editedat);
                channel.setCreator(authResult);
                if(result.channelprivate){
                    rows.forEach((row) => {
                        let user = {id: row.usersid, userName: row.username, firstName: row.firstname, lastName: row.lastname, photoURL: row.photourl};    
                        channel.pushMembers(user);
                    });                
                }
                resolve(channel);
            }
        });
    });
}

//TODO:Delete
//Handles DELETE /v1/channels/{channelID}
app.delete("/v1/channels/:channelID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if (authResult) {
        //checks if user is creator of channel
        db.query(Constants.SQL_GET_MEMBERS, [req.params.channelID], (err, rows) => {
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
    }
});

//Handles POST /v1/channels/{channelID}/members
app.post("/v1/channels/:channelID/members", async (req, res, next) => {    
    authResult = checkAuthentication(req, res);
    if(authResult){                
        try{
            //check user is creator of channel
            let creator = await checkIsCreator(req, authResult);
            if (!creator) {
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }            
            let channel = await getChannelMembers(req, authResult);
            //add user supplied in request body as member of this channel. respond 201 and text user was added
            //only id of user is required. (can post entire profile)
            if (req.body.id == null){
                return res.status(400).send("Should provide a user id to add to channel");
            }
            let addUser = req.body;
            query(db, Constants.SQL_INSERT_MEMBER + "(?, ?)", [req.params.channelID, addUser.id])
                .then((results) => {
                    console.log(results);
                })
                .catch(next);                                                                                                               
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(201).send("Member added to channel " + channel.name);  
        }catch(err) {
            next(err);
        }
    
    }
});

//TODO: check if private
//TODO: Don't delete creator?
//TODO: check rows affected?
//Handles DELETE /v1/channels/{channelID}/members
app.delete("/v1/channels/:channelID/members", async (req, res, next) => {  
    authResult = checkAuthentication(req, res);
    if(authResult){        
        //check user is creator of channel
        try{
            console.log("DELETE WORKS");
            //check user is creator of channel
            let creator = await checkIsCreator(req, authResult);
            if (!creator) {
                return res.status(403).send("Can't make changes to channel since you are not the creator");
            }            
            let channel = await getChannelMembers(req, authResult);

            if (req.body.id == null){
                return res.status(400).send("Should provide a user id to add to channel");
            }
            let addUser = req.body;
            query(db, Constants.SQL_DELETE_MEMBER, [req.params.channelID, addUser.id])
            .catch(next);
            res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
            res.status(200).send("Member deleted from channel " + channel.name);
        }catch(err) {
            next(err);
        }
            //remove user supplied in body from list of channel members
            //200 and text message indicate user removed.
            //only id required            
            
    }
});


//Handles PATCH /v1/messages/{messageID}
//change if not in channel anymore?
app.patch("/v1/messages/:messageID", (req, res, next) => {
    authResult = checkAuthentication(req, res);
    if(authResult){
        //check if user is creator of message
        db.query(Constants.SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
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
            db.query(Constants.SQL_UPDATE_MESSAGE, [newBody, req.params.messageID], (err, results) => {
                if (err) {
                    return next(err);
                }  
                db.query(Constants.SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
                    if (err) {
                        return next(err);
                    }     
                    message.setBody(rows[0].body);
                    res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_JSON);
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
        db.query(Constants.SQL_GET_MESSAGE, [req.params.messageID], (err, rows) => {
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
            db.query(Constants.SQL_DELETE_MESSAGE, [req.params.messageID], (err, results) => {
                if (err) {
                    return next(err);
                }  
                res.setHeader(Constants.CONTENT_TYPE, Constants.CONTENT_TEXT);
                res.status(200).send("Message deleted");
            });
        });
    }
});


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
    db.query(Constants.SQL_SELECT_CHANNEL, [channelID], (err, rows) => {
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