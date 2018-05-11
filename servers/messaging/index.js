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


const SQL_SELECT_ALL = "select id, channelname, channeldescription, channelprivate, createdat, creatorid, editedat from channel"

app.get("/v1/channels", (req, res, next) => {
    checkAuthentication(req, res);
    db.query(SQL_SELECT_ALL, (err, rows) => {
        if (err) {
            return next(err);
        }
        let channels = rows.map((row) => {
            //iterate
        })         
    })

});

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