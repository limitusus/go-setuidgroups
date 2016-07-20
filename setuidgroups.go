package main

import (
	"fmt"
	"os"
	"os/user"
	"log"
	"syscall"
	"strconv"
	"errors"
	"unsafe"
)

/*
#include <stdio.h>
#include <grp.h>
#include <errno.h>
static int mygetgrent(struct group** result, int* size, int* err) {
    struct group* g = getgrent();
    *size = 0;
    int r = 0;
    if (!g) {
        if (errno != 0) {
            *err = errno;
            r = errno;
        } else {
            *err = -1;
            r = -1;
        }
        *result = g;
        return r;
    }
    *result = g;
    char** p = g->gr_mem;
    //fprintf(stderr, "%s\n", g);
    //fprintf(stderr, "%s\n", g->gr_name);
    while (*p != NULL) {
        //fprintf(stderr, "  %s:%s\n", g->gr_name, *p);
        p++;
        (*size)++;
    }
    //fprintf(stderr, "name: %s", g->gr_name);
    return r;
}
*/
import "C"

type MyGroup struct {
	GrName string
	GrPasswd string
	GrGid int
	GrMembers []string
}

func main() {
	var args []string = os.Args
	var username = args[1]
	var prog = args[2:]
	user, err := user.Lookup(username)
	if err != nil {
		log.Fatal(err)
	}
	var supp_gids = make([]int, 0)
	for {
		group, err := getgrent()
		if err != nil {
			break
		}
		//dumpGroup(group)
		for i := 0; i < len(group.GrMembers); i++ {
			if group.GrMembers[i] == username {
				supp_gids = append(supp_gids, group.GrGid)
				break
			}
		}
	}
	/* Main switches */
	var uid, _ = strconv.Atoi(user.Uid)
	var gid, _ = strconv.Atoi(user.Gid)
	err = syscall.Setgid(gid)
	if err != nil {
		log.Fatal(err)
	}
	/*
	for i := 0; i < len(supp_gids); i++ {
		fmt.Printf("%d ", supp_gids[i])
	}
    */
	err = syscall.Setgroups(supp_gids)
	if err != nil {
		log.Fatal(err)
	}
	err = syscall.Setuid(uid)
	if err != nil {
		log.Fatal(err)
	}
	env := os.Environ()
	err = syscall.Exec(prog[0], prog[0:], env)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("should never come here\n")
}

func dumpGroup(g MyGroup) {
	fmt.Printf("GroupName: %s\n", g.GrName)
	fmt.Printf("  Passwd: %s\n", g.GrPasswd)
	fmt.Printf("  Gid: %d\n", g.GrGid)
	fmt.Printf("  Members: ")
	for i := 0; i < len(g.GrMembers); i++ {
		fmt.Printf("%s ", g.GrMembers[0])
	}
	fmt.Printf("\n")
}

func getgrent() (MyGroup, error) {
	var gent MyGroup
	var result *C.struct_group
	var err C.int
	var csize C.int
	var rv = C.mygetgrent(&result, &csize, &err)
	size := int(csize)
	if rv == -1 {
		return gent, errors.New("end")
	} else if rv != 0 {
		log.Fatal("getgrent failed: errno=" + strconv.Itoa(int(err)))
	}
	gent.GrName = C.GoString(result.gr_name)
	gent.GrPasswd = C.GoString(result.gr_passwd)
	gent.GrGid = int(result.gr_gid)
	gent.GrMembers = make([]string, 0)
	var cmembers **C.char = result.gr_mem
	slice := (*[1 << 30]*C.char)(unsafe.Pointer(cmembers))[:size:size]
	for i := 0; i < size; i++ {
		g := C.GoString(slice[i])
		gent.GrMembers = append(gent.GrMembers, g)
	}
	return gent, nil
}
