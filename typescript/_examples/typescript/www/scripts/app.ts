class User{
  private name: string;

  constructor(fullname:string) {
        this.name = fullname;
  }

  Hi(msg: string): string {
    return msg + " "+ this.name;
  }

}

var user = new User("siris web 22framework");
var hi = user.Hi("Hello");
window.alert(hi);
