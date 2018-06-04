module.exports = class Reaction {
    constructor(user, reaction){     
        this.user = user;
        this.reaction = reaction;
    }

    getUser() {
        return this.user;
    } 
    getReaction() {
        return this.reaction;
    }
    setUser(user) {
        this.user = user;
    }
    setReaction(reaction) {
        this.reaction = reaction;
    }

}