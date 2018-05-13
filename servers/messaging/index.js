"use strict";

import {Channel} from "./models/channel"
import {Message} from "./models/message"

const express = require("express");

const app = express();
const addr = process.env.ADDR || ":80";
const [host, port] = addr.split(":");
const portNum = parseInt(port);


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


const SQL_SELECT_ALL = "select id, channelname, channeldescription, channelprivate, createdat, creatorid, editedat from channel";
const SQL_SELECT_CHANNEL = "select * from channel_members cm " + 
                            "join channel c on c.id = cm.channelid " + 
                            "join members m on m.id = cm.membersid " + 
                            "where cm.channelid = ?";
const SQL_INSERT_CHANNEL = "insert into channel (channelname, channeldescription, channelprivate, createdat, creatorid, editedat) "+ 
                        "values (?, ?, ?, ?, ?, ?)";
const CONTENT_TYPE = "Content-Type";
const CONTENT_JSON = "application/json";

app.get("/v1/channels", (req, res, next) => {
    checkAuthentication(req, res);
    db.query(SQL_SELECT_ALL, (err, rows) => {
        if (err) {
            return next(err);
        }
        let channels = [];
        rows.forEach((row) => {            
            
            //method
            // db.query(SQL_SELECT_CHANNEL, [row.id], (err, rows) => {
            //     if (err) {
            //         return next(err);
            //     }  
            //     rows.forEach((row) => {
            //         members.push(row.userid)
            //     })              
            // });
            // channel.setMembers(getChannel(row.id));            
            channels.push(getChannel(row.id));
        }); 
        let channelsJSON = JSON.parse(channels);
        res.setHeader(CONTENT_TYPE, CONTENT_JSON);
        res.status(200).json(channelsJSON);       
    });    

});

app.post("v1/channels", (req, res, next) => {
    db.query(SQL_INSERT_CHANNEL, [req.body.name, req.body.description, req.body.private, req.body.createdat, 
        req.body.creatorid, req.body.editedat], (err, results) => {
            if (err) {
                return next(err);
            }
            let newID = results.insertId;            
            let channelJSON = JSON.parse(getChannel(newID));
            res.setHeader(CONTENT_TYPE, CONTENT_JSON);
            res.status(200).json(channelJSON);  
    });
})

app.use((err, req, res, next) => {
    if (err.stack) {
        console.error(err.stack);
    }
    res.status(500).send("Error in the server");
});

function checkAuthentication(req, res){
    let userJSON = req.get("X-User");
    if (userJSON) {
        let user = JSON.parse(userJSON);
        res.json(user);
    }else{
        res.status(401).send("Please sign in");
    }  
}

function getChannel(newID) {
    let channel;
    let members = [];
    db.query(SQL_SELECT_CHANNEL, [newID], (err, rows) => {
        if (err) {
            return next(err);
        }  
       channel = new Channel(rows[0].channelid, rows[0].channelname, rows[0].channeldescription, rows[0].channelprivate, rows[0].createdat, 
        rows[0].creatorid, rows[0].editedat);
        rows.forEach((row) => {
            members.push(row.userid)
        })   
        channel.setMembers(members)           
    });    
    return channel;
}