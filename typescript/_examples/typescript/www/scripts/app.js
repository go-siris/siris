class User {
    constructor(fullname) {
        this.name = fullname;
    }
    Hi(msg) {
        return msg + " " + this.name;
    }
}
var user = new User("siris web framework");
var hi = user.Hi("Hello");
window.alert(hi);
