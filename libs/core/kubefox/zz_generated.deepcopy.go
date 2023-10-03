//go:build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package kubefox

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Var) DeepCopyInto(out *Var) {
	*out = *in
	if in.arrayNumVal != nil {
		in, out := &in.arrayNumVal, &out.arrayNumVal
		*out = make([]float64, len(*in))
		copy(*out, *in)
	}
	if in.arrayStrVal != nil {
		in, out := &in.arrayStrVal, &out.arrayStrVal
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Var.
func (in *Var) DeepCopy() *Var {
	if in == nil {
		return nil
	}
	out := new(Var)
	in.DeepCopyInto(out)
	return out
}