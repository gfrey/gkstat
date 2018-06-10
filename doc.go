// G(o)KStat is a go library to access solaris kernel statistics (kstat). This
// is making heavy use of cgo and is hidden behind a build tag for solaris.
// Therefore no documentation is found here. A rough guideline on usage shall
// be given here:
//
//   ks, _ := Open()
//   defer ks.Close()
//
//   item, _ := ks.Find(FilterModule("cpu_stat"))
//   val, _ := item.UInt64("anonfree")
//
//   for _, item := range ks.Scan(FilterModule("cpu_stat")) {
//     val, _ := item.UInt64("anonfree")
//   }
//
package gkstat
